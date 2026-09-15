ALTER TABLE users
	MODIFY COLUMN email  VARCHAR(255) NULL,
	MODIFY COLUMN status ENUM('active','pending_review','suspended','deleted') NOT NULL DEFAULT 'active',
	ADD COLUMN middle_name            VARCHAR(100) DEFAULT NULL AFTER first_name,
	ADD COLUMN suffix                 VARCHAR(20)  DEFAULT NULL AFTER last_name,
	ADD COLUMN member_type            ENUM('regular','young_saver') NOT NULL DEFAULT 'regular' AFTER status,
	ADD COLUMN guardian_full_name     VARCHAR(150) DEFAULT NULL AFTER member_type,
	ADD COLUMN guardian_mobile_number VARCHAR(20)  DEFAULT NULL AFTER guardian_full_name,
	ADD COLUMN guardian_consent_at    TIMESTAMP    NULL DEFAULT NULL AFTER guardian_mobile_number;

-- Pre-registration OTP challenges (purpose='register') are created before
-- any user row exists, so user_id must be nullable here. Split into two
-- ALTER TABLE statements: MySQL rejects DROP FOREIGN KEY and ADD CONSTRAINT
-- of the same name within a single statement (error 1826), even though the
-- drop is logically ordered before the add.
ALTER TABLE otp_verifications
	DROP FOREIGN KEY fk_otp_user,
	MODIFY COLUMN user_id BIGINT UNSIGNED NULL;

ALTER TABLE otp_verifications
	ADD CONSTRAINT fk_otp_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
