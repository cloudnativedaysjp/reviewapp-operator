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

func newApplicationTemplate() *dreamkastv1alpha1.ApplicationTemplate {
	app := `apiVersion: argoproj.io/v1alpha1
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
      prune: true`

	return &dreamkastv1alpha1.ApplicationTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "applicationtemplate-test-ram",
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ApplicationTemplateSpec{
			StableTemplate:    app,
			CandidateTemplate: app,
		},
	}
}

func newManifestsTemplate() *dreamkastv1alpha1.ManifestsTemplate {
	kustomizationYaml := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: demo-dev-{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}
bases:
- ../../../base
- ./ns.yaml`
	nsYaml := `apiVersion: v1
kind: Namespace
metadata:
  name: demo-dev-{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}`
	m := make(map[string]string)
	m["kustomization.yaml"] = kustomizationYaml
	m["ns.yaml"] = nsYaml

	return &dreamkastv1alpha1.ManifestsTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manifeststemplate-test-ram",
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ManifestsTemplateSpec{
			StableData:    m,
			CandidateData: m,
		},
	}
}

func newReviewAppManager() *dreamkastv1alpha1.ReviewAppManager {
	return &dreamkastv1alpha1.ReviewAppManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ram",
			Namespace: testNamespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppManagerSpec{
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
				Message: `
* {{.AppRepo.Organization}}
* {{.AppRepo.Repository}}
* {{.AppRepo.PrNumber}}
* {{.InfraRepo.Organization}}
* {{.InfraRepo.Repository}}
* {{.Variables.AppRepositoryAlias}}
* {{.Variables.dummy}}`,
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
						Name:      "manifeststemplate-test-ram",
					}},
					Dirpath: "overlays/dev/{{.Variables.AppRepositoryAlias}}-{{.AppRepo.PrNumber}}",
				},
				ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
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
	}
}

func newReviewApp() *dreamkastv1alpha1.ReviewApp {
	return &dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ra-shotakitazawa-reviewapp-operator-demo-app-1",
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
				Message:              `message`,
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
					Dirpath: "overlays/dev/test-ra-1",
				},
				ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
					Template: dreamkastv1alpha1.NamespacedName{
						Namespace: testNamespace,
						Name:      "applicationtemplate-test-ra",
					},
					Filepath: ".apps/dev/test-ra-1.yaml",
				},
			},
			AppPrNum: testGitAppPrNumForRA,
			Application: `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: "test-ra-1"
  namespace: argocd
spec:
  project: "default"
  destination:
    namespace: "test-ra-1"
    server: "https://kubernetes.default.svc"
  source:
    repoURL: https://github.com/ShotaKitazawa/reviewapp-operator-demo-infra
    path: "overlays/dev/test-ra-1"
    targetRevision: master
  syncPolicy:
    automated:
      prune: true`,
			Manifests: map[string]string{
				"kustomization.yaml": `apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Kustomization
namespace: demo-dev-test-ra-1
bases:
- ../../../base
- ./ns.yaml`,
				"ns.yaml": `apiVersion: v1
kind: Namespace
metadata:
  name: demo-dev-test-ra-1`,
			},
		},
	}
}

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
