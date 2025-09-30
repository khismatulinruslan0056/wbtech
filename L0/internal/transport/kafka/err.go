package kafka

type NonRetriableError struct {
	Err error
}

func (e NonRetriableError) Error() string {
	return e.Err.Error()
}

func NewNonRetriableError(err error) NonRetriableError {
	return NonRetriableError{Err: err}
}
