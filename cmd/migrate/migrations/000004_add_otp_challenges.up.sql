CREATE TABLE otp_challenges (
	id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	reference_id    CHAR(64)        NOT NULL,
	user_id         BIGINT UNSIGNED NOT NULL,
	badge_number_id VARCHAR(50)     NOT NULL,
	otp_hash        CHAR(64)        NOT NULL,
	mobile_number   VARCHAR(20)     NOT NULL,
	purpose         VARCHAR(30)     NOT NULL DEFAULT 'login',
	attempts        TINYINT UNSIGNED NOT NULL DEFAULT 0,
	max_attempts    TINYINT UNSIGNED NOT NULL DEFAULT 5,
	expires_at      TIMESTAMP       NOT NULL,
	consumed_at     TIMESTAMP       NULL DEFAULT NULL,
	created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uq_otp_reference_id (reference_id),
	KEY idx_otp_user_id (user_id),
	CONSTRAINT fk_otp_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
