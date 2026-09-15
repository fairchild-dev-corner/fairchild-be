ALTER TABLE otp_verifications
	ADD COLUMN status     ENUM('pending','sent','failed') NOT NULL DEFAULT 'pending' AFTER otp_hash,
	ADD COLUMN send_error VARCHAR(255) DEFAULT NULL AFTER status;
