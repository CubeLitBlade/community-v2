package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProblemDetail struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

func writeProblem(c *gin.Context, status int, title string, detail string, code string) {
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

func writeInvalidRequest(c *gin.Context, detail string) {
	writeProblem(
		c,
		http.StatusBadRequest,
		"Invalid request",
		detail,
		"INVALID_REQUEST",
	)
}
