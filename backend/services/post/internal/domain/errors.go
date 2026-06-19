// Package domain represents the core business layer of the post service.
// It encapsulates domain-specific errors and acts as the root namespace for
// domain models and ports, ensuring the core remains independent of application
// and infrastructure concerns.
package domain

import "errors"

var (
	ErrIDInvalid        = errors.New("invalid id")
	ErrContentBlank     = errors.New("content is blank")
	ErrPermissionDenied = errors.New("permission denied")
)
