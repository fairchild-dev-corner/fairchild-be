package models

import "time"

// ContactMessage mirrors the `contact_messages` table - a message submitted
// through the public Contact Us form (see ContactForm.tsx on the frontend).
// Unlike account_claim_requests there's no approve/reject workflow; staff
// simply review submissions directly (admin dashboard, forthcoming - until
// then, a direct DB read), same as the claim-request review gap noted in
// AuthService.SubmitClaimRequestService.
type ContactMessage struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ContactMessageRequest is the request body for POST /contact.
type ContactMessageRequest struct {
	Name    string `json:"name" validate:"required,max=255"`
	Email   string `json:"email" validate:"required,email,max=255"`
	Message string `json:"message" validate:"required,max=2000"`
}

// NewsletterSubscription mirrors the `newsletter_subscriptions` table - an
// email captured via the "Get Notified" form (see AppPromo.tsx).
type NewsletterSubscription struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// NewsletterSubscribeRequest is the request body for POST /newsletter.
type NewsletterSubscribeRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}
