package ledger

import (
	"encoding/json"
	"errors"
	"io"
)

const (
	CodeInvalidArgument = "invalid_argument"
	CodeNotFound        = "not_found"
	CodeInternal        = "internal"
)

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data"`
	Error   *ErrorInfo `json:"error"`
}

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func NewInvalidArgumentError(message string) *AppError {
	return &AppError{Code: CodeInvalidArgument, Message: message}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{Code: CodeNotFound, Message: message}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, Err: err}
}

func SuccessResponse(data any) Response {
	return Response{
		Success: true,
		Data:    data,
		Error:   nil,
	}
}

func ErrorResponse(err error) Response {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return Response{
			Success: false,
			Data:    nil,
			Error: &ErrorInfo{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		}
	}

	return Response{
		Success: false,
		Data:    nil,
		Error: &ErrorInfo{
			Code:    CodeInternal,
			Message: "internal error",
		},
	}
}

func WriteResponse(w io.Writer, response Response) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}
