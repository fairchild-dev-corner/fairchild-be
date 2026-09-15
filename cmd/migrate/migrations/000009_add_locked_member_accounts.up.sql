CREATE TABLE locked_member_accounts (
	id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	user_id         BIGINT UNSIGNED  DEFAULT NULL,
	badge_number_id VARCHAR(50)      NOT NULL,
	ip_address      VARCHAR(45)      DEFAULT NULL,
	attempt_count   INT UNSIGNED     NOT NULL,
	reason          VARCHAR(255)     NOT NULL,
	is_locked       BOOLEAN          NOT NULL DEFAULT TRUE,
	locked_at       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	expires_at      TIMESTAMP        NOT NULL,
	notified_at     TIMESTAMP        NULL DEFAULT NULL,
	unlocked_at     TIMESTAMP        NULL DEFAULT NULL,
	created_at      TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY idx_locked_accounts_badge_active (badge_number_id, is_locked, expires_at),
	KEY idx_locked_accounts_user_id (user_id),
	CONSTRAINT fk_locked_account_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
