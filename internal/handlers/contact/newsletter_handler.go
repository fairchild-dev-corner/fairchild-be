package handlers

import (
	"net/http"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/contact"
	"fairchild_be/internal/utils"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

// handleSubscribeNewsletter records an email captured via the "Get
// Notified" form.
func (h *ContactHandler) handleSubscribeNewsletter(ctx *gin.Context) {
	var req models.NewsletterSubscribeRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	if err := h.contactService.SubscribeNewsletterService(ctx, &req); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_NEWSLETTER_SUBSCRIBED)
}
