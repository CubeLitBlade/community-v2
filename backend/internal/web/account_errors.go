package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/CubeLitBlade/community-v2/backend/internal/account"
)

var accountErrorProblems = []struct {
	err  error
	spec problemSpec
}{
	{
		err: account.ErrUsernameBlank,
		spec: problemSpec{
			status: http.StatusBadRequest,
			title:  "Username blank",
			code:   "ACCOUNT_USERNAME_BLANK",
			detail: "Username cannot be blank.",
		},
	},
	{
		err: account.ErrUsernameLength,
		spec: problemSpec{
			status: http.StatusBadRequest,
			title:  "Username bad length",
			code:   "ACCOUNT_USERNAME_BAD_LENGTH",
			detail: fmt.Sprintf(
				"Username must be between %d and %d characters.",
				account.MinUsernameLength,
				account.MaxUsernameLength,
			),
		},
	},
	{
		err: account.ErrPasswordEmpty,
		spec: problemSpec{
			status: http.StatusBadRequest,
			title:  "Password empty",
			code:   "ACCOUNT_PASSWORD_EMPTY",
			detail: "Password cannot be empty.",
		},
	},
	{
		err: account.ErrPasswordTooShort,
		spec: problemSpec{
			status: http.StatusBadRequest,
			title:  "Password too short",
			code:   "ACCOUNT_PASSWORD_TOO_SHORT",
			detail: fmt.Sprintf(
				"Password should be at least %d characters.",
				account.MinPasswordLength,
			),
		},
	},
	{
		err: account.ErrPasswordTooLong,
		spec: problemSpec{
			status: http.StatusBadRequest,
			title:  "Password too long",
			code:   "ACCOUNT_PASSWORD_TOO_LONG",
			detail: fmt.Sprintf(
				"Password should be no more than %d bytes.",
				account.MaxPasswordBytes,
			),
		},
	},
}

func accountProblem(err error) (problemSpec, bool) {
	for _, mapping := range accountErrorProblems {
		if errors.Is(err, mapping.err) {
			return mapping.spec, true
		}
	}

	return problemSpec{}, false
}
