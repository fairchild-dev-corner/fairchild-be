package cc

const (
	// Auth
	SUCCESS_REGISTER         = "account registered successfully"
	SUCCESS_LOGIN            = "logged in successfully"
	SUCCESS_SOCIAL_LOGIN     = "social login successful"
	SUCCESS_TOKEN_REFRESH    = "token refreshed successfully"
	SUCCESS_LOGOUT           = "logged out successfully"
	SUCCESS_OTP_SENT         = "otp sent successfully"
	SUCCESS_SESSIONS_FETCHED = "sessions fetched successfully"
	SUCCESS_PROFILE_UPDATED  = "profile updated successfully"
	SUCCESS_PASSWORD_CHANGED = "password changed successfully"
	SUCCESS_SETTINGS_FETCHED = "settings fetched successfully"
	SUCCESS_SETTINGS_UPDATED = "settings updated successfully"

	// Registration
	SUCCESS_REGISTER_OTP_SENT     = "registration otp sent successfully"
	SUCCESS_REGISTER_OTP_VERIFIED = "otp verified, you may now complete registration"
	SUCCESS_YOUNG_SAVER_REGISTER  = "young saver enrollment submitted, pending review"

	// Forgot Password
	SUCCESS_FORGOT_PASSWORD_OTP_SENT     = "password reset otp sent successfully"
	SUCCESS_FORGOT_PASSWORD_OTP_VERIFIED = "otp verified, you may now reset your password"
	SUCCESS_PASSWORD_RESET               = "password reset successfully"

	// Dev-only (see AuthHandler.RegisterDevRoutes)
	SUCCESS_ACCOUNT_UNLOCKED = "account unlocked successfully"

	// Account Claim Requests
	SUCCESS_CLAIM_REQUEST_SUBMITTED = "your request has been submitted for review"

	// Contact
	SUCCESS_CONTACT_MESSAGE_SUBMITTED = "your message has been sent, we'll get back to you soon"
	SUCCESS_NEWSLETTER_SUBSCRIBED     = "you're subscribed, we'll notify you when we launch"

	// Transactions
	SUCCESS_LAST_TRANSACTIONS_FETCHED = "last transactions fetched successfully"
	SUCCESS_LOAN_TRANSACTIONS_FETCHED = "loan transactions fetched successfully"

	// Health
	SUCCESS_COOP_HEALTH_FETCHED = "coop health score fetched successfully"

	// Profile
	SUCCESS_PROFILE_FETCHED = "profile fetched successfully"

	// Amortization
	SUCCESS_AMORTIZATION_FETCHED = "amortization schedule fetched successfully"

	// Loans
	SUCCESS_LOANS_FETCHED = "loans fetched successfully"

	// Accounts Receivable
	SUCCESS_AR_FETCHED = "accounts receivable fetched successfully"

	// Activity
	SUCCESS_ACTIVITY_FETCHED = "monthly activity fetched successfully"

	// Balance
	SUCCESS_BALANCE_FETCHED = "current balance fetched successfully"

	// Statement of Account
	SUCCESS_SOA_FETCHED = "statement of account fetched successfully"
)
