ALTER TABLE otp_verifications
	ADD COLUMN verified_at TIMESTAMP NULL DEFAULT NULL AFTER expires_at;
