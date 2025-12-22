package execution

type RetryableError struct {
	Err error
}

func (r RetryableError) Error() string {
	return r.Err.Error()
}

func IsRetryable(err error) bool {
	_, ok := err.(RetryableError)
	return ok
}
