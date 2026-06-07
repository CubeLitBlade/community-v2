// Package transport provides HTTP transport adapters for the authz domain.
package transport

import (
	"errors"
	"net/http"

	"github.com/cubelitblade/community-v2/backend/pkg/common/httperr"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/authz"
)

const (
	titleCheckFailed = "Authorization check failed"
)

const (
	codeCheckFailed = "AUTHZ_CHECK_FAILED"
)

func authzProblem(err error) (httperr.ProblemSpec, bool) {
	if errors.Is(err, authz.ErrCheckFailed) {
		return httperr.ProblemSpec{
			Status: http.StatusInternalServerError,
			Title:  titleCheckFailed,
			Code:   codeCheckFailed,
			Detail: detailCheckFailed(),
		}, true
	}

	//nolint:exhaustruct // Callers check the bool return before using the spec.
	return httperr.ProblemSpec{}, false
}

func detailCheckFailed() string {
	return "The authorization check could not be completed. Please try again later."
}

func detailInvalidCheckBody() string {
	return "Request body must be valid JSON and include user, relation, and object."
}
