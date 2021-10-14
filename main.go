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

package main

import (
	"flag"
	"os"
	"time"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/utils/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	resyncPeriod = time.Second * 30
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(dreamkastv1alpha1.AddToScheme(scheme))
	utilruntime.Must(argocd_application_v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
		//StacktraceLevel: zapcore.DPanicLevel,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		SyncPeriod:             &resyncPeriod,
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f5221b50.cloudnativedays.jp",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	{ // initialize controller: ReviewAppManager
		ramLogger := ctrl.Log.WithName("controllers").WithName("ReviewAppManager")
		gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(ramLogger)
		if err != nil {
			setupLog.Error(err, "unable to initialize", "service", "GitRemoteRepoApp")
			os.Exit(1)
		}
		if err = (&controllers.ReviewAppManagerReconciler{
			Client:                  mgr.GetClient(),
			Log:                     ramLogger,
			Scheme:                  mgr.GetScheme(),
			GitRemoteRepoAppService: gitRemoteRepoAppService,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ReviewAppManager")
			os.Exit(1)
		}
	}
	{ // initialize controller: ReviewApp
		raLogger := ctrl.Log.WithName("controllers").WithName("ReviewApp")
		gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(raLogger)
		if err != nil {
			setupLog.Error(err, "unable to initialize", "service", "GitRemoteRepoApp")
			os.Exit(1)
		}
		gitRemoteRepoInfraService, err := wire.NewGitRemoteRepoInfraService(raLogger, exec.New())
		if err != nil {
			setupLog.Error(err, "unable to initialize", "service", "GitRemoteRepoInfra")
			os.Exit(1)
		}
		if err = (&controllers.ReviewAppReconciler{
			Client:                    mgr.GetClient(),
			Log:                       raLogger,
			Scheme:                    mgr.GetScheme(),
			GitRemoteRepoAppService:   gitRemoteRepoAppService,
			GitRemoteRepoInfraService: gitRemoteRepoInfraService,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "ReviewApp")
			os.Exit(1)
		}
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
