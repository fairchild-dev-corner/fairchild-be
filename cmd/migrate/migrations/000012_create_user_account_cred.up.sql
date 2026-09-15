CREATE TABLE user_account_cred (
	id                      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	user_id                 BIGINT UNSIGNED NOT NULL,
	password_hash           VARCHAR(255) DEFAULT NULL,
	last_login_at           TIMESTAMP    NULL DEFAULT NULL,
	last_password_change_at TIMESTAMP    NULL DEFAULT NULL,
	created_at              TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at              TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uq_user_account_cred_user_id (user_id),
	CONSTRAINT fk_user_account_cred_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Backfill one cred row per existing user (password_hash/last_login_at carried
-- over as-is - bcrypt hashes keep working via password_service.go's format
-- detection) so every future lookup can INNER JOIN instead of LEFT JOIN.
INSERT INTO user_account_cred (user_id, password_hash, last_login_at, created_at)
SELECT id, password_hash, last_login_at, created_at FROM users;

ALTER TABLE users
	DROP COLUMN password_hash,
	DROP COLUMN last_login_at;
