package v1beta1

type NamespacedName struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
