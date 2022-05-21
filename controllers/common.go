package controllers

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	defaultResult = ctrl.Result{
		RequeueAfter: 30 * time.Second,
	}
)
