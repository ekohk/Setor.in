// Package response defines the standard API envelope { data, meta, error }
// and helpers to write it from a Gin handler.
package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// Envelope is the canonical response shape for ALL endpoints.
type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Meta  any        `json:"meta,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
}

// ErrorBody is the public error representation; never expose internal cause.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Pagination is the standard meta payload for paginated lists.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// OK writes 200 with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

// Created writes 201 with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data})
}

// NoContent writes 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated writes 200 with data + pagination meta.
func Paginated(c *gin.Context, data any, p Pagination) {
	if p.PageSize > 0 && p.Total > 0 {
		p.TotalPages = int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: p})
}

// Err writes the error using its mapped HTTP status.
// If the error is not an *AppError, it's converted to INTERNAL_ERROR (500).
func Err(c *gin.Context, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		slog.Error("unhandled error", "err", err.Error(), "path", c.FullPath())
		ae = apperr.Wrap(apperr.CodeInternal, "internal error", err)
	}

	if ae.HTTPStatus() >= 500 {
		slog.Error("server error",
			"code", ae.Code,
			"msg", ae.Message,
			"err", ae.Error(),
			"path", c.FullPath(),
		)
	}

	c.JSON(ae.HTTPStatus(), Envelope{
		Error: &ErrorBody{
			Code:    string(ae.Code),
			Message: ae.Message,
			Details: ae.Details,
		},
	})
}
