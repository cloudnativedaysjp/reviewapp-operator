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
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/glogr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

const ()

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
		gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(logger)
		Expect(err).ToNot(HaveOccurred())
		reconciler := ReviewAppManagerReconciler{
			Client:                  k8sClient,
			Scheme:                  scheme,
			Log:                     logger,
			Recorder:                recorder,
			GitRemoteRepoAppService: gitRemoteRepoAppService,
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
	It("should create ReviewApp when PR is opened", func() {
		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		ram, err := createSomeResourceForReviewAppManagerTest(ctx)
		Expect(err).NotTo(HaveOccurred())

		ra := dreamkastv1alpha1.ReviewApp{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ram-shotakitazawa-reviewapp-operator-demo-app-1"}, &ra)
		}).Should(Succeed())
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

	It("should delete ReviewApp when PR is closed", func() {
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

	It("should update status", func() {
		// Control external resources: open PR for test
		err := ghClient.OpenPr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
		Expect(err).NotTo(HaveOccurred())

		_, err = createSomeResourceForReviewAppManagerTest(ctx)
		Expect(err).NotTo(HaveOccurred())

		updated := dreamkastv1alpha1.ReviewAppManager{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "test-ram"}, &updated)
			if err != nil {
				return err
			}
			if len(updated.Status.SyncedPullRequests) == 0 {
				return errors.New("status should be updated")
			}
			return nil
		}).Should(Succeed())
	})
	//! [test]
})

func createSomeResourceForReviewAppManagerTest(ctx context.Context) (*dreamkastv1alpha1.ReviewAppManager, error) {
	at := newApplicationTemplate("applicationtemplate-test-ram")
	if err := k8sClient.Create(ctx, at); err != nil {
		return nil, err
	}
	mt := newManifestsTemplate("manifeststemplate-test-ram")
	if err := k8sClient.Create(ctx, mt); err != nil {
		return nil, err
	}
	ram := newReviewAppManager()
	if err := k8sClient.Create(ctx, ram); err != nil {
		return nil, err
	}
	return ram, nil
}
