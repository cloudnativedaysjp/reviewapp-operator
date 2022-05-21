//go:build integration_test
// +build integration_test

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/glogr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

var _ = Describe("ReviewAppManager controller", func() {
	//! [setup]
	var stopFunc func()

	BeforeEach(func() {
		// remove finalizers before delete resources
		raList := &dreamkastv1alpha1.ReviewAppList{}
		err := k8sClient.List(ctx, raList, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		for _, ra := range raList.Items {
			ra.Finalizers = nil
			err := k8sClient.Update(ctx, &ra)
			Expect(err).NotTo(HaveOccurred())
		}
		// delete resources
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ReviewAppManager{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ReviewApp{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ApplicationTemplate{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ManifestsTemplate{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())

		// Control external resources: close PR for test
		err = ghClient.ClosePr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		// initialize controller-manager of ReviewAppManager
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())
		logger := glogr.NewWithOptions(glogr.Options{LogCaller: glogr.None})
		k8sRepository, err := wire.NewKubernetes(logger, k8sClient)
		Expect(err).ToNot(HaveOccurred())
		gitApiRepository, err := wire.NewGitHubApi(logger)
		Expect(err).ToNot(HaveOccurred())
		reconciler := ReviewAppManagerReconciler{
			Scheme:   scheme,
			Log:      logger,
			Recorder: recorder,
			K8s:      k8sRepository,
			GitApi:   gitApiRepository,
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
		// remove finalizers before delete resources
		raList := &dreamkastv1alpha1.ReviewAppList{}
		err := k8sClient.List(ctx, raList, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		for _, ra := range raList.Items {
			ra.Finalizers = nil
			err := k8sClient.Update(ctx, &ra)
			Expect(err).NotTo(HaveOccurred())
		}
		// delete resources
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ReviewAppManager{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ReviewApp{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ApplicationTemplate{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &dreamkastv1alpha1.ManifestsTemplate{}, client.InNamespace(testNamespace))
		Expect(err).NotTo(HaveOccurred())

		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})
	//! [setup]

	//! [test]
	It("should be created ReviewApp when PR is opened", func() {
		// TODO: freeze time.Now()
		//now := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		ram, err := createSomeResourceForReviewAppManagerTest(ctx)
		Expect(err).NotTo(HaveOccurred())

		// wait to run reconcile loop
		ra := dreamkastv1alpha1.ReviewApp{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ram-shotakitazawa-reviewapp-operator-demo-app-1"}, &ra)
		},
			60*time.Second, // timeout
			10*time.Second, // interval
		).Should(Succeed())

		Expect(ra.Spec.AppTarget).To(Equal(ram.Spec.AppTarget))
		Expect(ra.Spec.InfraTarget).To(Equal(ram.Spec.InfraTarget))
		Expect(ra.Spec.AppConfig.Message).To(Equal(`
* ShotaKitazawa
* reviewapp-operator-demo-app
* 1
* ShotaKitazawa
* reviewapp-operator-demo-infra
* test-ram
* <no value>`))
	})

	// TODO
	// It("should be updated ReviewApp when PR is updated", func() {
	// })

	It("should be deleted ReviewApp when PR is closed", func() {
		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		_, err = createSomeResourceForReviewAppManagerTest(ctx)
		Expect(err).NotTo(HaveOccurred())

		// Control external resources: close PR for test
		err = ghClient.ClosePr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		// wait to run reconcile loop
		ra := dreamkastv1alpha1.ReviewApp{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ram-shotakitazawa-reviewapp-operator-demo-app-1"}, &ra)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return err
			}
			return fmt.Errorf("ReviewApp must not exist")
		},
			60*time.Second, // timeout
			10*time.Second, // interval
		).Should(Succeed())
	})

	It("should be created PullRequest when PR is opened", func() {
		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		_, err = createSomeResourceForReviewAppManagerTest(ctx)
		Expect(err).NotTo(HaveOccurred())

		pr := &dreamkastv1alpha1.PullRequest{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "shotakitazawa-reviewapp-operator-demo-app-1"}, pr)
			if err != nil {
				return err
			}
			return nil
		},
			60*time.Second, // timeout
			10*time.Second, // interval
		).Should(Succeed())

		Expect(pr.Spec.AppTarget.Organization).To(Equal(testGitAppOrganization))
		Expect(pr.Spec.AppTarget.Repository).To(Equal(testGitAppRepository))
		Expect(pr.Spec.AppTarget.Username).To(Equal(testGitUsername))
	})
	//! [test]
})

func createSomeResourceForReviewAppManagerTest(ctx context.Context) (*dreamkastv1alpha1.ReviewAppManager, error) {
	at := newApplicationTemplate("applicationtemplate-test-ram")
	if err := k8sClient.Create(ctx, at); err != nil {
		return nil, err
	}
	mt := newManifestsTemplate_RAM("manifeststemplate-test-ram")
	if err := k8sClient.Create(ctx, mt); err != nil {
		return nil, err
	}
	ram := newReviewAppManager_RAM()
	if err := k8sClient.Create(ctx, ram); err != nil {
		return nil, err
	}
	return ram, nil
}

func newReviewAppManager_RAM() *dreamkastv1alpha1.ReviewAppManager {
	return &dreamkastv1alpha1.ReviewAppManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ram",
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppManagerSpec{
			ReviewAppCommonSpec: dreamkastv1alpha1.ReviewAppCommonSpec{
				AppTarget: dreamkastv1alpha1.ReviewAppCommonSpecAppTarget{
					Username:     testGitUsername,
					Organization: testGitAppOrganization,
					Repository:   testGitAppRepository,
					GitSecretRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "git-creds",
						},
						Key: "token",
					},
					IgnoreLabels: []string{"ignore-reviewappmanager"},
				},
				AppConfig: dreamkastv1alpha1.ReviewAppCommonSpecAppConfig{
					Message: `
* {{.AppRepo.Organization}}
* {{.AppRepo.Repository}}
* {{.AppRepo.PrNumber}}
* {{.InfraRepo.Organization}}
* {{.InfraRepo.Repository}}
* {{.Variables.AppRepositoryAlias}}
* {{.Variables.dummy}}`,
				},
				InfraTarget: dreamkastv1alpha1.ReviewAppCommonSpecInfraTarget{
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
				InfraConfig: dreamkastv1alpha1.ReviewAppCommonSpecInfraConfig{
					Manifests: dreamkastv1alpha1.ReviewAppCommonSpecInfraManifests{
						Templates: []dreamkastv1alpha1.NamespacedName{{
							Namespace: testNamespace,
							Name:      "manifeststemplate-test-ram",
						}},
						Dirpath: "overlays/dev/{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}",
					},
					ArgoCDApp: dreamkastv1alpha1.ReviewAppCommonSpecInfraArgoCDApp{
						Template: dreamkastv1alpha1.NamespacedName{
							Namespace: testNamespace,
							Name:      "applicationtemplate-test-ram",
						},
						Filepath: ".apps/dev/{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}.yaml",
					},
				},
				Variables: []string{
					"AppRepositoryAlias=test-ram",
				},
			},
		},
	}
}

func newManifestsTemplate_RAM(name string) *dreamkastv1alpha1.ManifestsTemplate {
	kustomizationYaml := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}
bases:
- ../../../base
patchesStrategicMerge:
- ./manifests.yaml
`
	manifestsYaml := ` 
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo
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
        commit: {{.AppRepo.LatestCommitHash}}
    spec:
      containers:
        - name: demo
          image: nginx
`
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
