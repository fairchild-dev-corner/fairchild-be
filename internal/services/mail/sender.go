package mail

import "context"

// Sender delivers an HTML email to an address. Never trust delivery to have
// succeeded without checking the returned error - see SMTPClient.Send for
// provider-specific failure modes.
type Sender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}
