package handlers

import (
	"net/http"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/contact"
	"fairchild_be/internal/utils"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

// handleSubmitContactMessage stores a message submitted through the public
// Contact Us form for staff to review.
func (h *ContactHandler) handleSubmitContactMessage(ctx *gin.Context) {
	var req models.ContactMessageRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.contactService.SubmitContactMessageService(ctx, &req); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_CONTACT_MESSAGE_SUBMITTED)
}
