package transport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
	"github.com/CubeLitBlade/community-v2/backend/internal/httperr"
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
			Title:  "Invalid username",
			Code:   "ACCOUNT_USERNAME_BLANK",
			Detail: "Username cannot be blank.",
		},
	},
	{
		err: account.ErrUsernameLength,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  "Invalid username",
			Code:   "ACCOUNT_USERNAME_BAD_LENGTH",
			Detail: fmt.Sprintf(
				"Username must be between %d and %d characters.",
				account.MinUsernameLength,
				account.MaxUsernameLength,
			),
		},
	},
	{
		err: account.ErrPasswordEmpty,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  "Invalid password",
			Code:   "ACCOUNT_PASSWORD_EMPTY",
			Detail: "Password cannot be empty.",
		},
	},
	{
		err: account.ErrPasswordTooShort,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  "Invalid password",
			Code:   "ACCOUNT_PASSWORD_TOO_SHORT",
			Detail: fmt.Sprintf(
				"Password should be at least %d characters.",
				account.MinPasswordLength,
			),
		},
	},
	{
		err: account.ErrPasswordTooLong,
		spec: httperr.ProblemSpec{
			Status: http.StatusBadRequest,
			Title:  "Invalid password",
			Code:   "ACCOUNT_PASSWORD_TOO_LONG",
			Detail: fmt.Sprintf(
				"Password should be no more than %d bytes.",
				account.MaxPasswordBytes,
			),
		},
	},
	{
		err: account.ErrUsernameAlreadyExists,
		spec: httperr.ProblemSpec{
			Status: http.StatusConflict,
			Title:  "Username conflict",
			Code:   "ACCOUNT_USERNAME_ALREADY_EXISTS",
			Detail: "Username already exists.",
		},
	},
	{
		err: account.ErrInvalidCredentials,
		spec: httperr.ProblemSpec{
			Status: http.StatusUnauthorized,
			Title:  "Invalid credentials",
			Code:   "ACCOUNT_INVALID_CREDENTIALS",
			Detail: "Username or password is incorrect.",
		},
	},
}

func accountProblem(err error) (httperr.ProblemSpec, bool) {
	for _, mapping := range accountErrorProblems {
		if errors.Is(err, mapping.err) {
			return mapping.spec, true
		}
	}

	return httperr.ProblemSpec{
		Title:  "",
		Code:   "",
		Detail: "",
		Status: 0,
	}, false
}
