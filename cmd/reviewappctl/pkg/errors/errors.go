package errors

import "fmt"

type ErrorFilenameIsEmpty struct{}

func NewErrorFilenameIsEmpty() *ErrorFilenameIsEmpty {
	return &ErrorFilenameIsEmpty{}
}
func (e ErrorFilenameIsEmpty) Error() string {
	return fmt.Sprintf("filename is empty")
}

type ErrorFileNotFound struct {
	filename string
}

func NewErrorFileNotFound(filename string) *ErrorFileNotFound {
	return &ErrorFileNotFound{filename}
}

func (e ErrorFileNotFound) Error() string {
	return fmt.Sprintf("%s: no such files", e.filename)
}

type ErrorIsDirectory struct {
	filename string
}

func NewErrorIsDirectory(filename string) *ErrorIsDirectory {
	return &ErrorIsDirectory{filename}
}

func (e ErrorIsDirectory) Error() string {
	return fmt.Sprintf("%s is directory", e.filename)
}
