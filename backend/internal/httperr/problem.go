// Package httperr provides utilities for formatting and writing RFC 9457
// Problem Details HTTP responses.
package httperr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	titleBadRequest          = "Bad Request"
	titleUnauthorized        = "Unauthorized"
	titleInternalServerError = "Internal Server Error"
)

const (
	codeBadRequest          = "BAD_REQUEST"
	codeUnauthorized        = "UNAUTHORIZED"
	codeInternalServerError = "INTERNAL_SERVER_ERROR"
)

// Default error detail messages for common HTTP problem responses.
const (
	DefaultBadRequestMessage          = "The request parameters are invalid, please check and try again."
	DefaultUnauthorizedMessage        = "Please log in to access this resource."
	DefaultInternalServerErrorMessage = "An unexpected error occurred. Please try again later."
)

// ProblemDetail represents an RFC 9457 Problem Details object for HTTP APIs.
type ProblemDetail struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
	Status   int    `json:"status"`
}

// WriteProblem writes an RFC 9457 Problem Details JSON response to the Gin
// context.
func WriteProblem(c *gin.Context, status int, title, detail, code string) {
	c.Header("Content-Type", "application/problem+json")

	c.AbortWithStatusJSON(status, ProblemDetail{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.URL.Path,
		Code:     code,
	})
}

// WriteBadRequest writes a standard 400 Bad Request problem detail
// response.
func WriteBadRequest(c *gin.Context, detail string) {
	WriteProblem(c, http.StatusBadRequest, titleBadRequest, detail, codeBadRequest)
}

// WriteUnauthorized writes a standard 401 Unauthorized problem detail
func WriteUnauthorized(c *gin.Context, detail string) {
	WriteProblem(c, http.StatusUnauthorized, titleUnauthorized, detail, codeUnauthorized)
}

// WriteInternalServerError writes a standard 500 Internal Server Error problem detail
func WriteInternalServerError(c *gin.Context, detail string) {
	WriteProblem(c, http.StatusInternalServerError, titleInternalServerError,
		"", codeInternalServerError)
}

// ProblemSpec defines the specification for mapping an internal error to an HTTP problem.
type ProblemSpec struct {
	Title  string
	Code   string
	Detail string
	Status int
}

// WriteProblemSpec writes a problem detail response based on the provided ProblemSpec.
func WriteProblemSpec(c *gin.Context, spec ProblemSpec) {
	WriteProblem(c, spec.Status, spec.Title, spec.Detail, spec.Code)
}

// ErrorMapper is a function that attempts to map an error to a ProblemSpec.
type ErrorMapper func(error) (ProblemSpec, bool)

// WriteMappedError attempts to map an error using the provided mappers.
// If no mapper matches the error, it writes a generic 500 Internal Server Error.
func WriteMappedError(c *gin.Context, err error, mappers ...ErrorMapper) {
	for _, mapper := range mappers {
		if spec, ok := mapper(err); ok {
			WriteProblemSpec(c, spec)

			return
		}
	}

	WriteInternalServerError(c, DefaultInternalServerErrorMessage)
}
