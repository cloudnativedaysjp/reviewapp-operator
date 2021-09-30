package kubernetes

import (
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

var SkipSpecIsNotUpdatedPredicateForReviewAppManager = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldObject := e.ObjectOld.(*dreamkastv1beta1.ReviewAppManager)
		newObject := e.ObjectNew.(*dreamkastv1beta1.ReviewAppManager)
		// NO enqueue request if spec is not updated
		return !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
	},
}

var SkipSpecIsNotUpdatedPredicateForReviewApp = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldObject := e.ObjectOld.(*dreamkastv1beta1.ReviewApp)
		newObject := e.ObjectNew.(*dreamkastv1beta1.ReviewApp)
		// NO enqueue request if spec is not updated
		return !reflect.DeepEqual(oldObject.Spec, newObject.Spec)
	},
}
