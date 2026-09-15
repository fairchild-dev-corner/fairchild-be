package utils

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidatePayload binds the request body JSON into v and runs struct
// validation tags against it. It writes the error response itself and
// returns false when binding/validation fails - callers must return
// immediately in that case rather than continuing with a partially
// populated v.
func ValidatePayload(ctx *gin.Context, v interface{}) bool {
	// Read body once
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return false
	}

	// Reset body so it can be read again
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Bind JSON from buffered body
	if err := ctx.ShouldBindJSON(v); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return false
	}

	// Validate struct
	if err := Validate.Struct(v); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}

	return true
}
