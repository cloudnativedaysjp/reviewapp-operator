package testutils

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Dynamic struct {
	Client dynamic.Interface
	Mapper meta.RESTMapper
}

func InitDynamicClient(cfg *rest.Config) (*Dynamic, error) {
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	discoveryClient := c.Discovery()

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dynamicClient := dyn

	return &Dynamic{dynamicClient, mapper}, nil
}

func (d Dynamic) newClient(data []byte, obj *unstructured.Unstructured, ns string) (dynamic.ResourceInterface, error) {
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(data, nil, obj)
	if err != nil {
		return nil, err
	}

	mapping, err := d.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if ns != "" {
			obj.SetNamespace(ns)
		} else if obj.GetNamespace() == "" {
			obj.SetNamespace(metav1.NamespaceDefault)
		}
		return d.Client.Resource(mapping.Resource).Namespace(obj.GetNamespace()), nil
	} else {
		return d.Client.Resource(mapping.Resource), nil
	}
}

func (d Dynamic) CreateOrUpdate(data []byte, obj *unstructured.Unstructured, ns string) error {
	c, err := d.newClient(data, obj, ns)
	if err != nil {
		return err
	}
	if _, err := c.Patch(context.Background(), obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "test",
	}); err != nil {
		return err
	}
	return nil
}
