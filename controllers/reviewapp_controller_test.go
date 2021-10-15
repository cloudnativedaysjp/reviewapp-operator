//+build integration_test

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/glogr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/exec"
	"k8s.io/utils/pointer"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

const (
	testReviewappControllerName = "test-reviewapp"
)

var _ = Describe("ReviewApp controller", func() {
	fmt.Println(GinkgoParallelNode())
	//! [setup]
	var stopFunc func()

	BeforeEach(func() {
		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRA)
		Expect(err).NotTo(HaveOccurred())

		// initialize controller-manager of ReviewApp
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			SyncPeriod: &testResyncPeriod,
			Scheme:     scheme,
		})
		Expect(err).ToNot(HaveOccurred())
		logger := glogr.NewWithOptions(glogr.Options{LogCaller: glogr.None})
		gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(logger)
		Expect(err).ToNot(HaveOccurred())
		gitRemoteRepoInfraService, err := wire.NewGitRemoteRepoInfraService(logger, exec.New())
		Expect(err).ToNot(HaveOccurred())
		reconciler := ReviewAppReconciler{
			Client:                    k8sClient,
			Scheme:                    scheme,
			Log:                       logger,
			GitRemoteRepoAppService:   gitRemoteRepoAppService,
			GitRemoteRepoInfraService: gitRemoteRepoInfraService,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		// stop func
		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()

		time.Sleep(100 * time.Millisecond)
	})
	//! [setup]

	//! [test]
	timeout := 120 * time.Second
	interval := 10 * time.Second
	Context("step1. create ReviewApp", func() {
		It("should succeed to create ReviewApp", func() {
			_, err := createSomeResourceForReviewAppTest(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should comment to app-repo's PR when create ReviewApp", func() {
			Eventually(func(g Gomega) {
				// sync argocd application
				err := testutils.SyncArgoCDApplication(argocdCLIPath, "reviewapps")
				g.Expect(err).NotTo(HaveOccurred())
				// get latest message from PR
				msg, err := ghClient.GetLatestMessage(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRA)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(msg).To(Equal("message"))
			}, timeout, interval).Should(Succeed())
		})
		It("should update status", func() {
			Eventually(func(g Gomega) {
				// sync argocd application
				err := testutils.SyncArgoCDApplication(argocdCLIPath, "reviewapps")
				g.Expect(err).NotTo(HaveOccurred())
				// get status of RA
				ra := &dreamkastv1alpha1.ReviewApp{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-1"}, ra)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ra.Status.Sync.Status).To(Equal(dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo))
				g.Expect(ra.Status.Sync.ApplicationName).To(Equal("test-ra-1"))
				g.Expect(ra.Status.Sync.ApplicationNamespace).To(Equal("argocd"))
				g.Expect(ra.Status.Sync.AppRepoLatestCommitSha).NotTo(BeEmpty())
				g.Expect(ra.Status.Sync.InfraRepoLatestCommitSha).NotTo(BeEmpty())
			}, timeout, interval).Should(Succeed())
		})
	})
	It("should commit to infra-repo", func() {
		files, err := ghClient.GetUpdatedFilenamesInLatestCommit(testGitInfraOrganization, testGitInfraRepository, testGitInfraBranch)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(Equal([]string{
			".apps/dev/test-ra-1.yaml",
			"overlays/dev/test-ra-1/kustomization.yaml",
			"overlays/dev/test-ra-1/ns.yaml",
		}))
	})
	It("should check Argo CD Application", func() {
		argocdApp := &argocd_application_v1alpha1.Application{}
		err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-1"}, argocdApp)
		Expect(err).NotTo(HaveOccurred())
		Expect(argocdApp.Annotations[annotationAppOrgNameForArgoCDApplication]).To(Equal(testGitAppOrganization))
		Expect(argocdApp.Annotations[annotationAppRepoNameForArgoCDApplication]).To(Equal(testGitAppRepository))
		Expect(argocdApp.Annotations[annotationAppCommitHashForArgoCDApplication]).NotTo(BeEmpty())
	})
	Context("step2. apply ReviewApp", func() {
		It("should succeed to create ReviewApp", func() {
			_, err := updateSomeResourceForReviewAppTest(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should comment to app-repo's PR when create ReviewApp", func() {
			Eventually(func(g Gomega) {
				// sync argocd application
				err := testutils.SyncArgoCDApplication(argocdCLIPath, "reviewapps")
				g.Expect(err).NotTo(HaveOccurred())
				// get latest message from PR
				msg, err := ghClient.GetLatestMessage(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRA)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(msg).To(Equal("modified"))
			}, timeout, interval).Should(Succeed())
		})
		It("should update status", func() {
			ra := &dreamkastv1alpha1.ReviewApp{}
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-1"}, ra)
			Expect(err).NotTo(HaveOccurred())
			Expect(ra.Status.Sync.Status).To(Equal(dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo))
			Expect(ra.Status.Sync.ApplicationName).To(Equal("test-ra-1"))
			Expect(ra.Status.Sync.ApplicationNamespace).To(Equal("argocd"))
			Expect(ra.Status.Sync.AppRepoLatestCommitSha).NotTo(BeEmpty())
			Expect(ra.Status.Sync.InfraRepoLatestCommitSha).NotTo(BeEmpty())
		})
		It("should commit to infra-repo", func() {
			files, err := ghClient.GetUpdatedFilenamesInLatestCommit(testGitInfraOrganization, testGitInfraRepository, testGitInfraBranch)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(Equal([]string{
				"overlays/dev/test-ra-1/ns.yaml",
			}))
		})
		It("should check Argo CD Application", func() {
			argocdApp := &argocd_application_v1alpha1.Application{}
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-1"}, argocdApp)
			Expect(err).NotTo(HaveOccurred())
			Expect(argocdApp.Annotations[annotationAppOrgNameForArgoCDApplication]).To(Equal(testGitAppOrganization))
			Expect(argocdApp.Annotations[annotationAppRepoNameForArgoCDApplication]).To(Equal(testGitAppRepository))
			Expect(argocdApp.Annotations[annotationAppCommitHashForArgoCDApplication]).NotTo(BeEmpty())
		})
	})
	Context("step3. delete ReviewApp", func() {
		It("should succeed to delete ReviewApp", func() {
			err := k8sClient.Delete(context.Background(), &dreamkastv1alpha1.ReviewApp{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ra-shotakitazawa-reviewapp-operator-demo-app-1",
					Namespace: testNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})
		It("should delete Argo CD Application", func() {
			Eventually(func(g Gomega) error {
				// sync argocd application
				err := testutils.SyncArgoCDApplication(argocdCLIPath, "reviewapps")
				g.Expect(err).NotTo(HaveOccurred())
				// check
				argocdApp := argocd_application_v1alpha1.Application{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-1"}, &argocdApp)
				if err != nil {
					if apierrors.IsNotFound(err) {
						return nil
					}
					return err
				}
				return fmt.Errorf("Application must not exist")
			}, timeout, interval).Should(Succeed())
		})
		It("should commit to infra-repo", func() {
			files, err := ghClient.GetDeletedFilenamesInLatestCommit(testGitInfraOrganization, testGitInfraRepository, testGitInfraBranch)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(Equal([]string{
				".apps/dev/test-ra-1.yaml",
				"overlays/dev/test-ra-1/kustomization.yaml",
				"overlays/dev/test-ra-1/ns.yaml",
			}))
		})
	})
	//! [test]
})

func createSomeResourceForReviewAppTest(ctx context.Context) (*dreamkastv1alpha1.ReviewApp, error) {
	argoCDApp := newArgoCDApplication()
	if err := k8sClient.Create(context.Background(), argoCDApp); err != nil {
		return nil, err
	}
	ra := newReviewApp()
	if err := k8sClient.Create(ctx, ra); err != nil {
		return nil, err
	}
	return ra, nil
}

func updateSomeResourceForReviewAppTest(ctx context.Context) (*dreamkastv1alpha1.ReviewApp, error) {
	nsYaml := `apiVersion: v1
kind: Namespace
metadata:
  name: demo-dev-test-ra-1
  annotations:
    modified: "true"`

	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dreamkast.cloudnativedays.jp",
		Version: "v1alpha1",
		Kind:    "ReviewApp",
	})
	patch.SetNamespace(testNamespace)
	patch.SetName("test-ra-shotakitazawa-reviewapp-operator-demo-app-1")
	patch.UnstructuredContent()["spec"] = map[string]interface{}{
		"appRepoConfig": map[string]interface{}{
			"message": "modified",
		},
		"manifests": map[string]string{
			"ns.yaml": nsYaml,
		},
	}
	if err := k8sClient.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: testReviewappControllerName,
		Force:        pointer.Bool(true),
	}); err != nil {
		return nil, err
	}

	ra := dreamkastv1alpha1.ReviewApp{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-1"}, &ra); err != nil {
		return nil, err
	}
	return &ra, nil
}
