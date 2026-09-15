ALTER TABLE otp_verifications
	DROP FOREIGN KEY fk_otp_user,
	MODIFY COLUMN user_id BIGINT UNSIGNED NOT NULL;

ALTER TABLE otp_verifications
	ADD CONSTRAINT fk_otp_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE users
	DROP COLUMN guardian_consent_at,
	DROP COLUMN guardian_mobile_number,
	DROP COLUMN guardian_full_name,
	DROP COLUMN member_type,
	DROP COLUMN suffix,
	DROP COLUMN middle_name,
	MODIFY COLUMN status ENUM('active','suspended','deleted') NOT NULL DEFAULT 'active',
	MODIFY COLUMN email  VARCHAR(255) NOT NULL;
