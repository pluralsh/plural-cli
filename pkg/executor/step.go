package executor

type WrappedError struct {
	inner  error
	Output string
}

func (we *WrappedError) Error() string {
	return we.inner.Error()
}
