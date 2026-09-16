package models

import "time"

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	// UserStatusPendingReview is the initial status for a Young Saver
	// account - it exists (guardian consent + OTP already verified) but
	// hasn't been reviewed by the membership team yet, who flip it to
	// UserStatusActive once they have. It can still log in like any other
	// account (AuthService.LoginService only blocks UserStatusSuspended) -
	// restricting what a pending account can actually do once logged in is
	// a dashboard-side concern, not yet implemented.
	UserStatusPendingReview UserStatus = "pending_review"
	UserStatusSuspended     UserStatus = "suspended"
	UserStatusDeleted       UserStatus = "deleted"
	// UserStatusInactive is a 1:1 carryover of the legacy member_record.m_status
	// value from the old portal - it is not yet distinguished from
	// UserStatusActive by any login/access logic (AuthService.LoginService only
	// blocks UserStatusSuspended today).
	UserStatusInactive UserStatus = "inactive"
)

type MemberType string

const (
	MemberTypeRegular MemberType = "regular"
	// MemberTypeYoungSaver accounts are dependents enrolled by a parent or
	// guardian - see AuthService.RegisterYoungSaverService. They carry
	// Guardian* fields and start out UserStatusPendingReview.
	MemberTypeYoungSaver MemberType = "young_saver"
)

// User mirrors the `users` MySQL table. PasswordHash is nil for accounts
// created purely through social login (no local password set). Email is
// nullable because MemberTypeYoungSaver accounts may not have one (the
// guardian's contact info is what's required for that member type).
type User struct {
	ID                   int64      `json:"id"`
	UUID                 string     `json:"uuid"`
	Email                string     `json:"email,omitempty"`
	Username             *string    `json:"username,omitempty"`
	PasswordHash         *string    `json:"-"`
	FirstName            *string    `json:"first_name,omitempty"`
	MiddleName           *string    `json:"middle_name,omitempty"`
	LastName             *string    `json:"last_name,omitempty"`
	Suffix               *string    `json:"suffix,omitempty"`
	MemberID             *string    `json:"member_id,omitempty"`
	DateOfBirth          *string    `json:"date_of_birth,omitempty"`
	Gender               *string    `json:"gender,omitempty"`
	MobileNumber         *string    `json:"mobile_number,omitempty"`
	Location             *string    `json:"location,omitempty"`
	// Address is the member-editable mailing address set via
	// PATCH /auth/settings - distinct from the read-only legacy
	// client-master address GET /profile returns, which has no write path.
	Address              *string    `json:"address,omitempty"`
	DisplayName          *string    `json:"display_name,omitempty"`
	AvatarUrl            *string    `json:"avatar_url,omitempty"`
	Status               UserStatus `json:"status"`
	MemberType           MemberType `json:"member_type"`
	GuardianFullName     *string    `json:"guardian_full_name,omitempty"`
	GuardianMobileNumber *string    `json:"guardian_mobile_number,omitempty"`
	GuardianConsentAt    *time.Time `json:"guardian_consent_at,omitempty"`
	EmailVerifiedAt      *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty"`
	LastPasswordChangeAt *time.Time `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// UserResponse is the safe, outward-facing projection of User - it never
// carries the password hash.
type UserResponse struct {
	ID                   int64      `json:"id"`
	UUID                 string     `json:"uuid"`
	Email                string     `json:"email,omitempty"`
	Username             *string    `json:"username,omitempty"`
	FirstName            *string    `json:"first_name,omitempty"`
	MiddleName           *string    `json:"middle_name,omitempty"`
	LastName             *string    `json:"last_name,omitempty"`
	Suffix               *string    `json:"suffix,omitempty"`
	MemberID             *string    `json:"member_id,omitempty"`
	DateOfBirth          *string    `json:"date_of_birth,omitempty"`
	Gender               *string    `json:"gender,omitempty"`
	MobileNumber         *string    `json:"mobile_number,omitempty"`
	Location             *string    `json:"location,omitempty"`
	DisplayName          *string    `json:"display_name,omitempty"`
	AvatarUrl            *string    `json:"avatar_url,omitempty"`
	Status               UserStatus `json:"status"`
	MemberType           MemberType `json:"member_type"`
	GuardianFullName     *string    `json:"guardian_full_name,omitempty"`
	GuardianMobileNumber *string    `json:"guardian_mobile_number,omitempty"`
	GuardianConsentAt    *time.Time `json:"guardian_consent_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:                   u.ID,
		UUID:                 u.UUID,
		Email:                u.Email,
		Username:             u.Username,
		FirstName:            u.FirstName,
		MiddleName:           u.MiddleName,
		LastName:             u.LastName,
		Suffix:               u.Suffix,
		MemberID:             u.MemberID,
		DateOfBirth:          u.DateOfBirth,
		Gender:               u.Gender,
		MobileNumber:         u.MobileNumber,
		Location:             u.Location,
		DisplayName:          u.DisplayName,
		AvatarUrl:            u.AvatarUrl,
		Status:               u.Status,
		MemberType:           u.MemberType,
		GuardianFullName:     u.GuardianFullName,
		GuardianMobileNumber: u.GuardianMobileNumber,
		GuardianConsentAt:    u.GuardianConsentAt,
		CreatedAt:            u.CreatedAt,
	}
}
