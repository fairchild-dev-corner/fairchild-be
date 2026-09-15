ALTER TABLE users
	CHANGE COLUMN badge_number_id member_id VARCHAR(50) DEFAULT NULL;
ALTER TABLE users RENAME INDEX uq_users_badge_number_id TO uq_users_member_id;

ALTER TABLE otp_verifications
	CHANGE COLUMN badge_number_id member_id VARCHAR(50) NOT NULL;

ALTER TABLE locked_member_accounts
	CHANGE COLUMN badge_number_id member_id VARCHAR(50) NOT NULL;
ALTER TABLE locked_member_accounts RENAME INDEX idx_locked_accounts_badge_active TO idx_locked_accounts_member_active;

ALTER TABLE user_account_cred
	ADD COLUMN member_id VARCHAR(50) DEFAULT NULL AFTER user_id;

UPDATE user_account_cred c
JOIN users u ON u.id = c.user_id
SET c.member_id = u.member_id;

-- users.member_id keeps its UNIQUE key (renamed above), so it's a valid FK
-- target even though it's nullable - social/SSO-only accounts have no
-- member_id, and NULL is exempt from FK matching in InnoDB.
ALTER TABLE user_account_cred
	ADD CONSTRAINT fk_user_account_cred_member FOREIGN KEY (member_id) REFERENCES users(member_id) ON DELETE CASCADE;
