ALTER TABLE users
	ADD COLUMN password_hash VARCHAR(255) DEFAULT NULL AFTER username,
	ADD COLUMN last_login_at TIMESTAMP    NULL DEFAULT NULL AFTER email_verified_at;

UPDATE users u
JOIN user_account_cred c ON c.user_id = u.id
SET u.password_hash = c.password_hash,
    u.last_login_at = c.last_login_at;

DROP TABLE user_account_cred;
