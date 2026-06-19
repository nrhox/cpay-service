package errmsg

import "errors"

var (
	ErrOauthUnsupport         = errors.New("unsupport the provider")
	ErrOauthInvalidState      = errors.New("security token validation failed or session expired")
	ErrOauthhEmptyAuthCode    = errors.New("provider authorization code is empty")
	ErrOauthAuthProcessFailed = errors.New("authentication subsystem error")
	ErrOauthEmailNotVerify    = errors.New("email not verified")
	ErrInComplateUserRegister = errors.New("anda belum melengkapi formulir")
	ErrGithubApi              = errors.New("terjadi kesalahan saat masuk menggunakan github")
	ErrAccountSuspend         = errors.New("maaf akun anda ditangguhkan")
)
