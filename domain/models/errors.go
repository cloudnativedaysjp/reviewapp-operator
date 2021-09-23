package models

type K8sRsourceNotFound struct {
	Err error
}

func (e K8sRsourceNotFound) Error() string { return "" }

// utility functions

func IsNotFound(err error) bool {
	switch err.(type) {
	case K8sRsourceNotFound:
		return true
	default:
		return false
	}
}
