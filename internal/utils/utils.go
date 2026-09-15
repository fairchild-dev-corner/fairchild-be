package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	common "fairchild_be/internal/models/common"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, common.ErrorResponse{
		ErrorCode: status,
		TimeStamp: time.Now(),
		ErrorMsg:  err.Error(),
	})
}

func ParseJSON(r *gin.Context, v interface{}) error {
	if r.Request.Body == nil {
		return fmt.Errorf("invalid request body")
	}

	return json.NewDecoder(r.Request.Body).Decode(v)
}

func LoggerHandler(level slog.Level, message string, method string, url_path string, status int) {

	// Define a logger with a custom format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level.Level(),
	}))
	slog.SetDefault(logger)

	slog.LogAttrs(
		context.Background(),
		level,
		message,
		slog.Group("request",
			slog.String("method ", method),
			slog.String("url", url_path)),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
		slog.Int("HttpStatus", status),
	)
}

func MustStringSHA256(s string) string {

	h := sha256.New()
	h.Write([]byte(s))
	hashed := hex.EncodeToString(h.Sum(nil))

	return hashed
}
