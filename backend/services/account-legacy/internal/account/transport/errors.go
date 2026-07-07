package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/cubelitblade/community-v2/backend/services/account-legacy/internal/account"
)

//nolint:gochecknoglobals // read-only static mapping used for error translation
const (
	titleAccountNotFound      = "Account not found"
	titleAuthenticationFailed = "Invalid credentials"
	titleInvalidPassword      = "Invalid password"
	titleInvalidUsername      = "Invalid username"
	titleUsernameConflict     = "Username conflict"
)

const (
	codeAccountNotFound       = "ACCOUNT_NOT_FOUND	"
	codeAuthenticationFailed  = "ACCOUNT_INVALID_CREDENTIALS"
	codePasswordEmpty         = "ACCOUNT_PASSWORD_EMPTY"
	codePasswordTooLong       = "ACCOUNT_PASSWORD_TOO_LONG"
	codePasswordTooShort      = "ACCOUNT_PASSWORD_TOO_SHORT"
	codeUsernameAlreadyExists = "ACCOUNT_USERNAME_ALREADY_EXISTS"
	codeUsernameBadLength     = "ACCOUNT_USERNAME_BAD_LENGTH"
	codeUsernameBlank         = "ACCOUNT_USERNAME_BLANK"
)

//nolint:gochecknoglobals // read-only static mapping used for error translation
var accountErrorProblems = []struct {
	err  error
	spec httperr.ProblemSpec
}{
	{
		err: account.ErrUsernameBlank,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  titleInvalidUsername,
			Code:   codeUsernameBlank,
			Detail: detailUsernameBlank(),
		},
	},
	{
		err: account.ErrUsernameLength,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  titleInvalidUsername,
			Code:   codeUsernameBadLength,
			Detail: detailUsernameBadLength(),
		},
	},
	{
		err: account.ErrPasswordEmpty,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  titleInvalidPassword,
			Code:   codePasswordEmpty,
			Detail: detailPasswordEmpty(),
		},
	},
	{
		err: account.ErrPasswordTooShort,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  titleInvalidPassword,
			Code:   codePasswordTooShort,
			Detail: detailPasswordTooShort(),
		},
	},
	{
		err: account.ErrPasswordTooLong,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  titleInvalidPassword,
			Code:   codePasswordTooLong,
			Detail: detailPasswordTooLong(),
		},
	},
	{
		err: account.ErrUsernameAlreadyExists,
		spec: httperr.ProblemSpec{
			Status: http.StatusConflict,
			Title:  titleUsernameConflict,
			Code:   codeUsernameAlreadyExists,
			Detail: detailUsernameAlreadyExists(),
		},
	},
	{
		err: account.ErrInvalidCredentials,
		spec: httperr.ProblemSpec{
			Status: http.StatusUnauthorized,
			Title:  titleAuthenticationFailed,
			Code:   codeAuthenticationFailed,
			Detail: detailInvalidCredentials(),
		},
	},
	{
		err: account.ErrAccountNotFound,
		spec: httperr.ProblemSpec{
			Status: http.StatusNotFound,
			Title:  titleAccountNotFound,
			Code:   codeAccountNotFound,
			Detail: detailAccountNotFound(),
		},
	},
}

func accountProblem(err error) (httperr.ProblemSpec, bool) {
	for _, mapping := range accountErrorProblems {
		if errors.Is(err, mapping.err) {
			return mapping.spec, true
		}
	}

	//nolint:exhaustruct // Callers check the bool return before using the spec.
	return httperr.ProblemSpec{}, false
}

func detailUsernameBlank() string {
	return "Username cannot be blank."
}

func detailUsernameBadLength() string {
	return fmt.Sprintf(
		"Username must be between %d and %d characters.",
		account.MinUsernameLength,
		account.MaxUsernameLength,
	)
}

func detailPasswordEmpty() string {
	return "Password cannot be empty."
}

func detailPasswordTooShort() string {
	return fmt.Sprintf(
		"Password should be at least %d characters.",
		account.MinPasswordLength,
	)
}

func detailPasswordTooLong() string {
	return fmt.Sprintf(
		"Password should be no more than %d bytes.",
		account.MaxPasswordBytes,
	)
}

func detailUsernameAlreadyExists() string {
	return "Username already exists."
}

func detailInvalidCredentials() string {
	return "Username or password is incorrect."
}

func detailAccountNotFound() string {
	return "Account not found."
}

func detailInvalidCreateAccountBody() string {
	return "Request body must be valid JSON and include username and password."
}
