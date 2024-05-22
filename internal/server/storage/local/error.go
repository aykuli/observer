package local

import (
	"fmt"
)

type fileStorageErr struct {
	name       string
	methodName string
	Err        error
}

func (fsErr *fileStorageErr) Error() string {
	return fmt.Sprintf("[%s]: method [%s]: %s", fsErr.name, fsErr.methodName, fsErr.Err.Error())
}

func newFSError(methodName string, err error) error {
	return &fileStorageErr{
		name:       "File Storage",
		methodName: methodName,
		Err:        err,
	}
}
