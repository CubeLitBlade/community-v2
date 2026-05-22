package transport

import (
	"errors"
	"net/http"

	"github.com/cubelitblade/community-v2/backend/services/user/internal/auth"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
)

//nolint:gochecknoglobals // read-only static mapping used for error translation
var authErrorProblems = []struct {
	err  error
	spec httperr.ProblemSpec
}{
	{
		err: auth.ErrInvalidCredentials,
		spec: httperr.ProblemSpec{
			Status: http.StatusUnauthorized,
			Title:  "Invalid credentials",
			Code:   "AUTH_INVALID_CREDENTIALS",
			Detail: "Username or password is incorrect",
		},
	},
}

func authProblem(err error) (httperr.ProblemSpec, bool) {
	for _, mapping := range authErrorProblems {
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
