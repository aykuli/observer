package ram

import (
	"fmt"
)

type ramStorageErr struct {
	name       string
	methodName string
	Err        error
}

func (rsErr *ramStorageErr) Error() string {
	return fmt.Sprintf("[%s]: method [%s]: %s", rsErr.name, rsErr.methodName, rsErr.Err.Error())
}

func newRAMErr(methodName string, err error) error {
	return &ramStorageErr{
		name:       "RAM Storage",
		methodName: methodName,
		Err:        err,
	}
}
