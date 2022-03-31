//go:build integration_test
// +build integration_test

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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		k8sRepository, err := wire.NewKubernetesRepository(logger, k8sClient)
		Expect(err).ToNot(HaveOccurred())
		gitApiRepository, err := wire.NewGitHubAPIRepository(logger)
		Expect(err).ToNot(HaveOccurred())
		gitCommandRepository, err := wire.NewGitCommandRepository(logger, exec.New())
		Expect(err).ToNot(HaveOccurred())
		reconciler := ReviewAppReconciler{
			Scheme:               scheme,
			Log:                  logger,
			Recorder:             recorder,
			K8sRepository:        k8sRepository,
			GitApiRepository:     gitApiRepository,
			GitCommandRepository: gitCommandRepository,
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
			argoCDApp := newArgoCDApplication()
			err := k8sClient.Create(context.Background(), argoCDApp)
			Expect(err).NotTo(HaveOccurred())
			at := newApplicationTemplate("applicationtemplate-test-ra")
			err = k8sClient.Create(ctx, at)
			Expect(err).NotTo(HaveOccurred())
			mt := newManifestsTemplate("manifeststemplate-test-ra", 1)
			err = k8sClient.Create(ctx, mt)
			Expect(err).NotTo(HaveOccurred())
			ra := newReviewApp("test-ra-shotakitazawa-reviewapp-operator-demo-app-2", 1)
			err = k8sClient.Create(ctx, ra)
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
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-2"}, ra)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ra.Status.Sync.Status).To(Equal(dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates))
				g.Expect(ra.Status.Sync.ApplicationName).To(Equal("test-ra-2"))
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
			".apps/dev/test-ra-2.yaml",
			"overlays/dev/test-ra-2/kustomization.yaml",
			"overlays/dev/test-ra-2/manifests.yaml",
		}))
	})
	It("should check Argo CD Application", func() {
		argocdApp := &argocd_application_v1alpha1.Application{}
		err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-2"}, argocdApp)
		Expect(err).NotTo(HaveOccurred())
		Expect(argocdApp.Annotations[annotationAppOrgNameForArgoCDApplication]).To(Equal(testGitAppOrganization))
		Expect(argocdApp.Annotations[annotationAppRepoNameForArgoCDApplication]).To(Equal(testGitAppRepository))
		Expect(argocdApp.Annotations[annotationAppCommitHashForArgoCDApplication]).NotTo(BeEmpty())
	})
	Context("step2. update ReviewApp", func() {
		It("should succeed to create ReviewApp", func() {
			mt := newManifestsTemplate("manifeststemplate-test-ra", 2)
			err := k8sClient.Patch(ctx, mt, client.Apply, &client.PatchOptions{
				FieldManager: testReviewappControllerName,
				Force:        pointer.Bool(true),
			})
			Expect(err).NotTo(HaveOccurred())
			ra := newReviewApp("test-ra-shotakitazawa-reviewapp-operator-demo-app-2", 2)
			err = k8sClient.Patch(ctx, ra, client.Apply, &client.PatchOptions{
				FieldManager: testReviewappControllerName,
				Force:        pointer.Bool(true),
			})
			Expect(err).NotTo(HaveOccurred())
			ra = &dreamkastv1alpha1.ReviewApp{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-2"}, ra)
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
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-shotakitazawa-reviewapp-operator-demo-app-2"}, ra)
			Expect(err).NotTo(HaveOccurred())
			Expect(ra.Status.Sync.Status).To(Equal(dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates))
			Expect(ra.Status.Sync.ApplicationName).To(Equal("test-ra-2"))
			Expect(ra.Status.Sync.ApplicationNamespace).To(Equal("argocd"))
			Expect(ra.Status.Sync.AppRepoLatestCommitSha).NotTo(BeEmpty())
			Expect(ra.Status.Sync.InfraRepoLatestCommitSha).NotTo(BeEmpty())
		})
		It("should commit to infra-repo", func() {
			files, err := ghClient.GetUpdatedFilenamesInLatestCommit(testGitInfraOrganization, testGitInfraRepository, testGitInfraBranch)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(Equal([]string{
				"overlays/dev/test-ra-2/manifests.yaml",
			}))
		})
		It("should check Argo CD Application", func() {
			argocdApp := &argocd_application_v1alpha1.Application{}
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-2"}, argocdApp)
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
					Name:      "test-ra-shotakitazawa-reviewapp-operator-demo-app-2",
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
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ra-2"}, &argocdApp)
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
				".apps/dev/test-ra-2.yaml",
				"overlays/dev/test-ra-2/kustomization.yaml",
				"overlays/dev/test-ra-2/manifests.yaml",
			}))
		})
	})
	//! [test]
})

//! [constructors for test]
func newArgoCDApplication() *argocd_application_v1alpha1.Application {
	return &argocd_application_v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "reviewapps",
			Namespace: "argocd",
		},
		Spec: argocd_application_v1alpha1.ApplicationSpec{
			Project: "default",
			Destination: argocd_application_v1alpha1.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "argocd",
			},
			Source: argocd_application_v1alpha1.ApplicationSource{
				RepoURL:        "https://github.com/ShotaKitazawa/reviewapp-operator-demo-infra",
				Path:           ".apps/dev",
				TargetRevision: "master",
				Directory: &argocd_application_v1alpha1.ApplicationSourceDirectory{
					Recurse: true,
				},
			},
			SyncPolicy: &argocd_application_v1alpha1.SyncPolicy{
				Automated: &argocd_application_v1alpha1.SyncPolicyAutomated{
					Prune: true,
				},
			},
		},
	}
}

func newApplicationTemplate(name string) *dreamkastv1alpha1.ApplicationTemplate {
	app := `
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: "{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}"
  namespace: argocd
spec:
  project: "default"
  destination:
    namespace: "{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}"
    server: "https://kubernetes.default.svc"
  source:
    repoURL: https://github.com/ShotaKitazawa/reviewapp-operator-demo-infra
    path: "overlays/dev/{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}"
    targetRevision: master
  syncPolicy:
    automated:
      prune: true
    syncOptions:
    - CreateNamespace=true
`
	return &dreamkastv1alpha1.ApplicationTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ApplicationTemplateSpec{
			StableTemplate:    app,
			CandidateTemplate: app,
		},
	}
}

func newManifestsTemplate(name string, step int) *dreamkastv1alpha1.ManifestsTemplate {
	kustomizationYaml := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}
bases:
- ../../../base
patchesStrategicMerge:
- ./manifests.yaml
`
	manifestsYaml := fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
  annotations:
    step: "%d"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
      annotations:
        commit: {{.AppRepo.LatestCommitSha}}
    spec:
      containers:
        - name: demo
          image: nginx
`, step)
	m := make(map[string]string)
	m["kustomization.yaml"] = kustomizationYaml
	m["manifests.yaml"] = manifestsYaml

	return &dreamkastv1alpha1.ManifestsTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ManifestsTemplateSpec{
			StableData:    m,
			CandidateData: m,
		},
	}
}

func newReviewApp(objectName string, step int) *dreamkastv1alpha1.ReviewApp {
	return &dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppSpec{
			AppTarget: dreamkastv1alpha1.ReviewAppManagerSpecAppTarget{
				Username:     testGitUsername,
				Organization: testGitAppOrganization,
				Repository:   testGitAppRepository,
				GitSecretRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "git-creds",
					},
					Key: "token",
				},
			},
			AppConfig: dreamkastv1alpha1.ReviewAppManagerSpecAppConfig{
				Message:              fmt.Sprintf("step %d", step),
				SendMessageEveryTime: true,
			},
			InfraTarget: dreamkastv1alpha1.ReviewAppManagerSpecInfraTarget{
				Username:     testGitUsername,
				Organization: testGitInfraOrganization,
				Repository:   testGitInfraRepository,
				Branch:       testGitInfraBranch,
				GitSecretRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "git-creds",
					},
					Key: "token",
				},
			},
			InfraConfig: dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig{
				Manifests: dreamkastv1alpha1.ReviewAppManagerSpecInfraManifests{
					Templates: []dreamkastv1alpha1.NamespacedName{{
						Namespace: testNamespace,
						Name:      "manifeststemplate-test-ra",
					}},
					Dirpath: "overlays/dev/test-ra-2",
				},
				ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
					Template: dreamkastv1alpha1.NamespacedName{
						Namespace: testNamespace,
						Name:      "applicationtemplate-test-ra",
					},
					Filepath: ".apps/dev/test-ra-2.yaml",
				},
			},
			Variables: []string{
				"AppRepositoryAlias=test-ra",
			},
			AppPrNum: testGitAppPrNumForRA,
		},
	}
}

//! [constructors for test]
