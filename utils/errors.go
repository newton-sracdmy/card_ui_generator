package utils

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(code int, message string, err error) error {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewBadRequestError(message string, err error) error {
	return NewError(400, message, err)
}

func NewUnprocessableEntityError(message string, err error) error {
	return NewError(422, message, err)
}

func NewConflictError(message string, err error) error {
	return NewError(409, message, err)
}

func NewNotFoundError(message string, err error) error {
	return NewError(404, message, err)
}

func NewInternalServerError(message string, err error) error {
	return NewError(500, message, err)
}

func NewServiceUnavailableError(message string, err error) error {
	return NewError(503, message, err)
}


func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return 500
}
