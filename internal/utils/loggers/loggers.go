package loggers

import (
	cc "fairchild_be/internal/constants"
	common "fairchild_be/internal/models/common"
	"fairchild_be/internal/utils"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/**
* Logs Definition
* @Description: Define the logging statement here
**/

func StatusOK(ctx *gin.Context, data any) {
	utils.LoggerHandler(
		slog.LevelInfo,
		fmt.Sprintf("%v", data),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusOK,
	)

	ctx.JSON(http.StatusOK, data)
}

func StatusNoContent(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusNoContent,
	)

	ctx.JSON(http.StatusNoContent, err)
}

func StatusUnauthorizedError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusUnauthorized,
	)

	ctx.JSON(http.StatusUnauthorized, err)
}

func StatusInternalServerError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusInternalServerError,
	)

	ctx.JSON(http.StatusInternalServerError, err)
}

func StatusBadRequestError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusBadRequest,
	)

	ctx.JSON(http.StatusBadRequest, err)
}

func StatusNotFound(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusNotFound,
	)

	ctx.JSON(http.StatusNotFound, err)
}

func StatusConflictError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusConflict,
	)

	ctx.JSON(http.StatusConflict, err)
}

func StatusTooManyRequestsError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusTooManyRequests,
	)

	ctx.JSON(http.StatusTooManyRequests, err)
}

func StatusLockedError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusLocked,
	)

	ctx.JSON(http.StatusLocked, err)
}

func StatusForbiddenError(ctx *gin.Context, err any) {
	utils.LoggerHandler(
		slog.LevelError,
		fmt.Sprintf("%v", err),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		http.StatusForbidden,
	)

	ctx.JSON(http.StatusForbidden, err)
}

// ErrorHandler captures errors and returns a consistent JSON error response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Step1: Process the request first.

		// Step2: Check if any errors were added to the context
		if len(c.Errors) > 0 {
			// Step3: Use the last error
			err := c.Errors.Last().Err

			errCode := http.StatusInternalServerError
			// Step4: Respond with a generic error message
			c.JSON(errCode, gin.H{
				"errCode": errCode,
				"message": err.Error(),
			})
		}
		// Any other steps if no errors are found
	}
}

func GetCommonError(ctx *gin.Context, err string, errCode int) {

	r := &common.ErrorResponse{
		ErrorCode: errCode,
		TimeStamp: time.Now(),
		ErrorMsg:  err,
	}

	switch errCode {
	case http.StatusNoContent:
		StatusNoContent(ctx, r)
	case http.StatusBadRequest:
		StatusBadRequestError(ctx, r)
	case http.StatusInternalServerError:
		StatusInternalServerError(ctx, r)
	case http.StatusUnauthorized:
		StatusUnauthorizedError(ctx, r)
	case http.StatusNotFound:
		StatusNotFound(ctx, r)
	case http.StatusConflict:
		StatusConflictError(ctx, r)
	case http.StatusTooManyRequests:
		StatusTooManyRequestsError(ctx, r)
	case http.StatusLocked:
		StatusLockedError(ctx, r)
	case http.StatusForbidden:
		StatusForbiddenError(ctx, r)
	default:
		// A status code not wired into this switch must still write SOME
		// response - silently returning here leaves gin to send an empty
		// 200 OK, which a client-side success path then tries to parse as
		// a normal payload (e.g. a login response body of "undefined"),
		// crashing the caller far from this actual cause.
		slog.Error("unhandled error status code in GetCommonError, falling back to 500", "code", errCode)
		StatusInternalServerError(ctx, r)
	}
}

func SuccessResponse(ctx *gin.Context, data any) {
	StatusOK(ctx, &common.SuccessResponse{
		SuccessID:  uuid.NewString(),
		Status:     cc.SUCCESS,
		HttpCode:   http.StatusOK,
		ResponseAt: time.Now(),
		Body:       data,
	})
}
