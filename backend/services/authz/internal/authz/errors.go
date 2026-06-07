package authz

import "errors"

// ErrCheckFailed is returned when the OpenFGA check request fails.
var ErrCheckFailed = errors.New("authorization check failed")
