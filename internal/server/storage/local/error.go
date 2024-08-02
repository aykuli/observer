package local

type fileStorageErr struct {
	name       string
	methodName string
	Err        error
}

func (fsErr *fileStorageErr) Error() string {
	return "[" + fsErr.name + "]: method [" + fsErr.methodName + "]: " + fsErr.Err.Error()
}

func newFSError(methodName string, err error) error {
	return &fileStorageErr{
		name:       "File Storage",
		methodName: methodName,
		Err:        err,
	}
}
