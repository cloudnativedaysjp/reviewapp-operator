package kubernetes

import (
	"context"

	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

func GetNamespacedNameFromObjectStr(ctx context.Context, objectStr string) (types.NamespacedName, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(objectStr), &obj)
	if err != nil {
		return types.NamespacedName{}, xerrors.Errorf("%w", err)
	}
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, nil
}

func SetAnnotationToObjectStr(ctx context.Context, objectStr string, annotationKey, annotationValue string) (string, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(objectStr), &obj)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	m := obj.GetAnnotations()
	if m == nil {
		m = make(map[string]string)
	}
	m[annotationKey] = annotationValue
	obj.SetAnnotations(m)

	b, err := yaml.Marshal(&obj)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return string(b), nil
}
