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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
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
		gitRemoteRepoInfraService, err := wire.NewGitRemoteRepoInfraService(logger)
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
			_, err := createOrSkipSomeResourceForReviewAppTest(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should update status", func() {
			Eventually(func(g Gomega) error {
				// sync argocd application
				err := testutils.SyncArgoCDApplication(argocdCLIPath, "reviewapps")
				g.Expect(err).NotTo(HaveOccurred())
				// get k8s-object of argocd application
				ra := &dreamkastv1beta1.ReviewApp{}
				if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-1"}, ra); err != nil {
					return err
				}
				g.Expect(ra.Status.Sync.Status).To(Equal(dreamkastv1beta1.SyncStatusCodeWatchingAppRepo))
				g.Expect(ra.Status.Sync.ApplicationName).To(Equal("test-ra-1"))
				g.Expect(ra.Status.Sync.ApplicationNamespace).To(Equal("argocd"))
				g.Expect(ra.Status.Sync.AppRepoLatestCommitSha).NotTo(BeEmpty())
				g.Expect(ra.Status.Sync.InfraRepoLatestCommitSha).NotTo(BeEmpty())
				return nil
			}, timeout, interval).Should(Succeed())
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
			Expect(argocdApp.Annotations[kubernetes.AnnotationAppOrgNameForArgoCDApplication]).To(Equal(testGitAppOrganization))
			Expect(argocdApp.Annotations[kubernetes.AnnotationAppRepoNameForArgoCDApplication]).To(Equal(testGitAppRepository))
			Expect(argocdApp.Annotations[kubernetes.AnnotationAppCommitHashForArgoCDApplication]).NotTo(BeEmpty())
		})
		It("should comment to app-repo's PR when create ReviewApp", func() {
			msg, err := ghClient.GetLatestMessage(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRA)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(Equal("message"))
		})
	})
	Context("step2. apply ReviewApp", func() {
		// TODO
	})
	Context("step3. delete ReviewApp", func() {
		// TODO
	})
	//! [test]
})

func createOrSkipSomeResourceForReviewAppTest(ctx context.Context, step int) (*dreamkastv1beta1.ReviewApp, error) {
	secret := newSecret()
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		secret := newSecret()
		if err := k8sClient.Create(ctx, secret); err != nil {
			return nil, err
		}
	}
	argoCDApp := newArgoCDApplication()
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: argoCDApp.Namespace, Name: argoCDApp.Name}, argoCDApp); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		argoCDApp := newArgoCDApplication()
		if err := k8sClient.Create(context.Background(), argoCDApp); err != nil {
			return nil, err
		}
	}
	ra := newReviewApp()
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: ra.Namespace, Name: ra.Name}, ra); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		switch step {
		case 1:
			ra = newReviewApp()
		case 2:
			ra = newReviewAppStep2()
		default:
			return nil, fmt.Errorf("unknown step")
		}
		if err := k8sClient.Create(ctx, ra); err != nil {
			return nil, err
		}
	}
	return ra, nil
}
