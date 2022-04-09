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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	scheme    = runtime.NewScheme()
	recorder  = record.NewFakeRecorder(1000)
	ctx       = context.Background()

	gitCredential string
	ghClient      *testutils.GitHubClient
	argocdCLIPath string
	kustomizePath string
)

const (
	testGitUsername          = "ShotaKitazawa"
	testGitAppOrganization   = "ShotaKitazawa"
	testGitAppRepository     = "reviewapp-operator-demo-app"
	testGitAppPrNumForRAM    = 1
	testGitAppPrNumForRA     = 2
	testGitInfraOrganization = "ShotaKitazawa"
	testGitInfraRepository   = "reviewapp-operator-demo-infra"
	testGitInfraBranch       = "master"
	testNamespace            = "argocd"
)

var (
	testResyncPeriod = time.Second * 5
)

func init() {
	gitCredential = os.Getenv("TEST_GITHUB_TOKEN")
	if gitCredential == "" {
		fmt.Println("environment variable TEST_GITHUB_TOKEN is not set")
		os.Exit(1)
	}
	ghClient = testutils.NewGitHubClient(gitCredential)

	argocdCLIPath = os.Getenv("ARGOCD_CLI_PATH")
	if argocdCLIPath == "" {
		argocdCLIPath = "/tmp/.reviewapp-operator/argocd"
	}
	kustomizePath = os.Getenv("KUSTOMIZE_PATH")
	if kustomizePath == "" {
		kustomizePath = "/tmp/.reviewapp-operator/kustomize"
	}
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	var err error

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = dreamkastv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = argocd_application_v1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	// initialize k8s client
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// initalize k8s dynamic client
	dynamic, err := testutils.InitDynamicClient(cfg)
	Expect(err).NotTo(HaveOccurred())

	// install Argo CD
	manifest, err := testutils.KustomizeBuildForTest(kustomizePath)
	Expect(err).NotTo(HaveOccurred())
	Eventually(func(g Gomega) {
		// install
		decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(manifest), 100)
		for {
			var rawObj runtime.RawExtension
			if err := decoder.Decode(&rawObj); err != nil {
				break
			}
			obj := &unstructured.Unstructured{}
			err = dynamic.CreateOrUpdate(rawObj.Raw, obj, testNamespace)
			g.Expect(err).NotTo(HaveOccurred())
		}
		// check argocd-server is up
		var deployment appsv1.Deployment
		err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "argocd-server"}, &deployment)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(deployment.Status.ReadyReplicas).NotTo(Equal(int32(0)))
		g.Expect(deployment.Status.UnavailableReplicas).To(Equal(int32(0)))
	}, 180, 10).Should(Succeed())

	// install credential of github
	secret := newSecret()
	err = k8sClient.Create(ctx, secret)
	Expect(err).NotTo(HaveOccurred())
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

	// Delete K8s resources
	secret := newSecret()
	err = k8sClient.Delete(ctx, secret)
	Expect(err).NotTo(HaveOccurred())

	// Control external resources: close PR for test
	err = ghClient.ClosePr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRAM)
	Expect(err).NotTo(HaveOccurred())
	err = ghClient.ClosePr(testGitAppOrganization, testGitAppRepository, testGitAppPrNumForRA)
	Expect(err).NotTo(HaveOccurred())
})

func newSecret() *corev1.Secret {
	m := make(map[string]string)
	m["token"] = gitCredential
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "git-creds",
			Namespace: testNamespace,
		},
		StringData: m,
	}
}
