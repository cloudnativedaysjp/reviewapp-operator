package errors

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

/* K8sObjectNotFound */

type K8sObjectNotFound struct {
	Err             error
	ObjectGVK       schema.GroupVersionKind
	ObjectName      string
	ObjectNamespace string
}

func NewK8sObjectNotFound(err error, gvk schema.GroupVersionKind, nn types.NamespacedName) K8sObjectNotFound {
	return K8sObjectNotFound{
		Err:             err,
		ObjectGVK:       gvk,
		ObjectName:      nn.Name,
		ObjectNamespace: nn.Namespace,
	}
}

func (e K8sObjectNotFound) Error() string {
	return fmt.Sprintf("%s %s/%s not found", e.ObjectGVK.Kind, e.ObjectNamespace, e.ObjectName)
}

func IsNotFound(err error) bool {
	switch err.(type) {
	case K8sObjectNotFound:
		return true
	default:
		return false
	}
}

/* KeyIsMissing */

type KeyIsMissing struct {
	kind string
	key  string
}

func NewKeyIsMissing(kind, key string) KeyIsMissing {
	return KeyIsMissing{kind, key}
}

func (e KeyIsMissing) Error() string {
	return fmt.Sprintf("in %s: key %s is missing", e.kind, e.key)
}

func IsKeyMissing(err error) bool {
	switch err.(type) {
	case KeyIsMissing:
		return true
	default:
		return false
	}
}
