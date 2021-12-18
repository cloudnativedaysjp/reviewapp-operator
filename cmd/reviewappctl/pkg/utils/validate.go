package utils

import (
	"os"

	"github.com/cloudnativedaysjp/reviewapp-operator/cmd/reviewappctl/pkg/errors"
)

func ValidateFile(filename string) error {
	if filename == "" {
		return errors.NewErrorFilenameIsEmpty()
	}
	if f, err := os.Stat(filename); err != nil {
		return errors.NewErrorFileNotFound(filename)
	} else if f.IsDir() {
		return errors.NewErrorIsDirectory(filename)
	}
	return nil
}
