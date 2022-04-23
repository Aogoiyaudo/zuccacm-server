package handler

var (
	ErrBadRequest  = ErrorMessage{"请求参数错误"}
	ErrNotLogged   = ErrorMessage{"未登录，请先登录"}
	ErrLoginFailed = ErrorMessage{"登录失败，SSO认证失败"}
	ErrForbidden   = ErrorMessage{"权限不足"}
	ErrNotFound    = ErrorMessage{"资源不存在"}
)

type ErrorMessage struct {
	Msg string
}

func (e ErrorMessage) Error() string {
	return e.Msg
}
