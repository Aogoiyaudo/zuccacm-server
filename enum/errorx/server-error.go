package errorx

import (
	"errors"
	"net/http"
)

type ErrorType int

const (
	ErrBadRequest = ErrorType(iota)
	ErrNotLogged
	ErrLoginFailed
	ErrForbidden
	ErrNotFound
)

var errMsg = map[ErrorType]string{
	ErrBadRequest:  "请求参数错误",
	ErrNotLogged:   "未登录，请先登录",
	ErrLoginFailed: "登录失败，SSO认证失败",
	ErrForbidden:   "权限不足",
	ErrNotFound:    "资源不存在",
}

func (t ErrorType) New() CustomError {
	return CustomError{
		errorType:     t,
		originalError: nil,
	}
}

func (t ErrorType) WithMessage(msg string) CustomError {
	return CustomError{
		errorType:     t,
		originalError: errors.New(msg),
	}
}

func (t ErrorType) Wrap(err error) CustomError {
	return CustomError{
		errorType:     t,
		originalError: err,
	}
}

type CustomError struct {
	errorType     ErrorType
	originalError error
}

func (e CustomError) Error() string {
	return errMsg[e.errorType]
}

func (e CustomError) Cause() error {
	return e.originalError
}

func (e CustomError) StatusCode() (code int) {
	switch e.errorType {
	case ErrBadRequest:
		code = http.StatusBadRequest
	case ErrNotLogged, ErrLoginFailed:
		code = http.StatusUnauthorized
	case ErrNotFound:
		code = http.StatusNotFound
	case ErrForbidden:
		code = http.StatusForbidden
	}
	return
}
