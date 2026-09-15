package models

import "time"

type ClaimRequestStatus string

const (
	ClaimRequestStatusPending  ClaimRequestStatus = "pending"
	ClaimRequestStatusApproved ClaimRequestStatus = "approved"
	ClaimRequestStatusRejected ClaimRequestStatus = "rejected"
)

// ClaimRequest mirrors the `account_claim_requests` table - a self-submitted
// request from a member with no email/mobile on file (so Forgot Password
// can't reach them, e.g. a legacy account migrated without portal
// credentials - see cmd/migrate_legacy_members) to attach contact info to
// their account. Never applied automatically: a staff member must verify
// the member's identity out-of-band and approve it before ProposedEmail/
// ProposedMobileNumber get copied onto the users row - see
// AuthService.SubmitClaimRequestService.
type ClaimRequest struct {
	ID                   int64              `json:"id"`
	MemberID             string             `json:"member_id"`
	FullName             *string            `json:"full_name,omitempty"`
	ProposedEmail        *string            `json:"proposed_email,omitempty"`
	ProposedMobileNumber *string            `json:"proposed_mobile_number,omitempty"`
	Message              *string            `json:"message,omitempty"`
	Status               ClaimRequestStatus `json:"status"`
	RejectionReason      *string            `json:"rejection_reason,omitempty"`
	ReviewedAt           *time.Time         `json:"reviewed_at,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
}

// ClaimAccountRequest is the request body for POST /auth/account-claim.
type ClaimAccountRequest struct {
	MemberID             string `json:"member_id" validate:"required,max=50"`
	FullName             string `json:"full_name" validate:"required,max=255"`
	ProposedEmail        string `json:"proposed_email" validate:"required,email"`
	ProposedMobileNumber string `json:"proposed_mobile_number" validate:"required"`
	Message              string `json:"message" validate:"omitempty,max=1000"`
}
