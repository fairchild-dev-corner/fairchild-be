package sms

import "fmt"

const brandName = "FCCMPC"

// LoginOTPMessage is sent for the second factor of login - see
// AuthService.createAndSendLoginOTP.
func LoginOTPMessage(code string, expiryMinutes int) string {
	return fmt.Sprintf("Your %s verification code is %s. It expires in %d minutes. Do not share this code with anyone.", brandName, code, expiryMinutes)
}

// ForgotPasswordOTPMessage is for the password-reset flow.
func ForgotPasswordOTPMessage(code string, expiryMinutes int) string {
	return fmt.Sprintf("Your %s password reset code is %s. It expires in %d minutes. If you didn't request this, you can ignore this message.", brandName, code, expiryMinutes)
}

// RegistrationOTPMessage is for verifying a mobile number during registration.
func RegistrationOTPMessage(code string, expiryMinutes int) string {
	return fmt.Sprintf("Your %s registration verification code is %s. It expires in %d minutes. Do not share this code with anyone.", brandName, code, expiryMinutes)
}
