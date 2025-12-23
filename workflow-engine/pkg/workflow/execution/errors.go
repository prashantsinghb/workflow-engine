package execution

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RetryableError struct {
	Err error
}

func (r RetryableError) Error() string {
	return r.Err.Error()
}

func IsRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
