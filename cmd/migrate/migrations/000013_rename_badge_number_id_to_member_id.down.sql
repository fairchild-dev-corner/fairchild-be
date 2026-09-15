ALTER TABLE user_account_cred DROP FOREIGN KEY fk_user_account_cred_member;
ALTER TABLE user_account_cred DROP COLUMN member_id;

ALTER TABLE locked_member_accounts RENAME INDEX idx_locked_accounts_member_active TO idx_locked_accounts_badge_active;
ALTER TABLE locked_member_accounts
	CHANGE COLUMN member_id badge_number_id VARCHAR(50) NOT NULL;

ALTER TABLE otp_verifications
	CHANGE COLUMN member_id badge_number_id VARCHAR(50) NOT NULL;

ALTER TABLE users RENAME INDEX uq_users_member_id TO uq_users_badge_number_id;
ALTER TABLE users
	CHANGE COLUMN member_id badge_number_id VARCHAR(50) DEFAULT NULL;
