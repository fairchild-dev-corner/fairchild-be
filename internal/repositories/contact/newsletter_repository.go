package repository

import (
	"github.com/gin-gonic/gin"
)

// CreateNewsletterSubscription inserts a new newsletter_subscriptions row.
// Re-subscribing with an email already on file is a silent no-op (the
// ON DUPLICATE KEY UPDATE is a no-op update, not an insert) rather than an
// error, so the response can't be used to enumerate which emails are
// already subscribed.
func (r *ContactRepository) CreateNewsletterSubscription(ctx *gin.Context, email string) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO newsletter_subscriptions (email) VALUES (?) ON DUPLICATE KEY UPDATE email = email`,
		email,
	)
	return err
}
