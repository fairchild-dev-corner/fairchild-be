package mail

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/otp_verification.html
var templateFS embed.FS

//go:embed templates/account_locked.html
var accountLockedFS embed.FS

var otpVerificationTmpl = template.Must(template.ParseFS(templateFS, "templates/otp_verification.html"))

var accountLockedTmpl = template.Must(template.ParseFS(accountLockedFS, "templates/account_locked.html"))

const brandName = "FCCMPC"

// OTPVerificationSubject is the subject line paired with
// RenderOTPVerificationEmail's body.
const OTPVerificationSubject = brandName + " Verification Code"

// OTPEmailData fills the placeholders in templates/otp_verification.html.
// Fields are inserted via html/template, which auto-escapes them - Code is
// server-generated (see GenerateOTPCode) and thus not attacker-controlled,
// but escaping applies regardless.
type OTPEmailData struct {
	BrandName     string
	Code          string
	ExpiryMinutes int
}

// RenderOTPVerificationEmail renders the OTP verification HTML email body
// (login, password reset, or registration - the copy is generic enough to
// cover all three; use OTPVerificationSubject as its Subject).
func RenderOTPVerificationEmail(code string, expiryMinutes int) (string, error) {
	var buf bytes.Buffer
	data := OTPEmailData{BrandName: brandName, Code: code, ExpiryMinutes: expiryMinutes}
	if err := otpVerificationTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render otp verification email: %w", err)
	}
	return buf.String(), nil
}

// AccountLockedSubject is the subject line paired with
// RenderAccountLockedEmail's body.
const AccountLockedSubject = brandName + " Account Locked"

// AccountLockedEmailData fills the placeholders in
// templates/account_locked.html.
type AccountLockedEmailData struct {
	BrandName  string
	LockedDays int
}

// RenderAccountLockedEmail renders the email sent when a member's account is
// locked out after too many failed login attempts - see
// AuthService.lockMemberAccount.
func RenderAccountLockedEmail(lockedDays int) (string, error) {
	var buf bytes.Buffer
	data := AccountLockedEmailData{BrandName: brandName, LockedDays: lockedDays}
	if err := accountLockedTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render account locked email: %w", err)
	}
	return buf.String(), nil
}
