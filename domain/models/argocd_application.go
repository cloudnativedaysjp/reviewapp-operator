package models

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ArgoCDApplication struct {
	metav1.ObjectMeta
	// TODO
	Status string
}
