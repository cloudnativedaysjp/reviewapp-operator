package errors

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type K8sObjectNotFound struct {
	Err             error
	ObjectGVK       schema.GroupVersionKind
	ObjectName      string
	ObjectNamespace string
}

// obj はオブジェクトの取得に失敗しているので、client.Object ではなく runtime.Object を利用する
func NewK8sObjectNotFound(err error, obj runtime.Object, nn types.NamespacedName) K8sObjectNotFound {
	return K8sObjectNotFound{
		Err:             err,
		ObjectGVK:       obj.GetObjectKind().GroupVersionKind(),
		ObjectName:      nn.Name,
		ObjectNamespace: nn.Namespace,
	}
}

func (e K8sObjectNotFound) Error() string {
	return fmt.Sprintf("%s %s/%s not found", e.ObjectGVK.Kind, e.ObjectNamespace, e.ObjectName)
}

// utility functions

func IsNotFound(err error) bool {
	switch err.(type) {
	case K8sObjectNotFound:
		return true
	default:
		return false
	}
}
