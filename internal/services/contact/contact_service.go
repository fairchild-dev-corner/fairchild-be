package services

import (
	models "fairchild_be/internal/models/contact"
	repository "fairchild_be/internal/repositories/contact"

	"github.com/gin-gonic/gin"
)

type ContactService struct {
	repo *repository.ContactRepository
}

func NewContactService(repo *repository.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

// SubmitContactMessageService records a message submitted through the public
// Contact Us form. Always succeeds from the caller's point of view - there's
// no applicant identity to verify up front (unlike account claim requests),
// so any validated payload is stored for staff to read.
func (s *ContactService) SubmitContactMessageService(ctx *gin.Context, req *models.ContactMessageRequest) error {
	_, err := s.repo.CreateContactMessage(ctx, &models.ContactMessage{
		Name:    req.Name,
		Email:   req.Email,
		Message: req.Message,
	})
	return err
}

// SubscribeNewsletterService records an email captured via the "Get
// Notified" form. Always succeeds from the caller's point of view, including
// for an email that's already subscribed - see
// ContactRepository.CreateNewsletterSubscription.
func (s *ContactService) SubscribeNewsletterService(ctx *gin.Context, req *models.NewsletterSubscribeRequest) error {
	return s.repo.CreateNewsletterSubscription(ctx, req.Email)
}
