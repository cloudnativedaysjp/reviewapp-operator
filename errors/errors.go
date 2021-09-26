package errors

type K8sResourceNotFound struct {
	Err error
}

func (e K8sResourceNotFound) Error() string { return "" }

// utility functions

func IsNotFound(err error) bool {
	switch err.(type) {
	case K8sResourceNotFound:
		return true
	default:
		return false
	}
}
