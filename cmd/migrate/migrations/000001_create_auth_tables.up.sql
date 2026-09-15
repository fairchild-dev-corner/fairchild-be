CREATE TABLE auth_providers (
	id         TINYINT UNSIGNED NOT NULL AUTO_INCREMENT,
	code       VARCHAR(30)      NOT NULL,
	name       VARCHAR(50)      NOT NULL,
	is_enabled BOOLEAN          NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uq_auth_providers_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE users (
	id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	uuid              CHAR(36)        NOT NULL,
	email             VARCHAR(255)    NOT NULL,
	username          VARCHAR(50)     DEFAULT NULL,
	password_hash     VARCHAR(255)    DEFAULT NULL,
	display_name      VARCHAR(150)    DEFAULT NULL,
	avatar_url        VARCHAR(512)    DEFAULT NULL,
	status            ENUM('active','suspended','deleted') NOT NULL DEFAULT 'active',
	email_verified_at TIMESTAMP       NULL DEFAULT NULL,
	last_login_at     TIMESTAMP       NULL DEFAULT NULL,
	created_at        TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at        TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at        TIMESTAMP       NULL DEFAULT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY uq_users_uuid (uuid),
	UNIQUE KEY uq_users_email (email),
	UNIQUE KEY uq_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_social_accounts (
	id               BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
	user_id          BIGINT UNSIGNED  NOT NULL,
	provider_id      TINYINT UNSIGNED NOT NULL,
	provider_user_id VARCHAR(191)     NOT NULL,
	provider_email   VARCHAR(255)     DEFAULT NULL,
	access_token     TEXT             DEFAULT NULL,
	refresh_token    TEXT             DEFAULT NULL,
	token_expires_at TIMESTAMP        NULL DEFAULT NULL,
	raw_profile      JSON             DEFAULT NULL,
	created_at       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at       TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uq_provider_account (provider_id, provider_user_id),
	KEY idx_social_user_id (user_id),
	CONSTRAINT fk_social_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_social_provider FOREIGN KEY (provider_id) REFERENCES auth_providers(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE auth_sessions (
	id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	user_id            BIGINT UNSIGNED NOT NULL,
	refresh_token_hash CHAR(64)        NOT NULL,
	user_agent         VARCHAR(255)    DEFAULT NULL,
	ip_address         VARCHAR(45)     DEFAULT NULL,
	expires_at         TIMESTAMP       NOT NULL,
	revoked_at         TIMESTAMP       NULL DEFAULT NULL,
	created_at         TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY uq_session_token_hash (refresh_token_hash),
	KEY idx_session_user_id (user_id),
	CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE login_audit_logs (
	id              BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
	user_id         BIGINT UNSIGNED  DEFAULT NULL,
	provider_id     TINYINT UNSIGNED DEFAULT NULL,
	email_attempted VARCHAR(255)     DEFAULT NULL,
	ip_address      VARCHAR(45)      DEFAULT NULL,
	user_agent      VARCHAR(255)     DEFAULT NULL,
	status          ENUM('success','failed') NOT NULL,
	reason          VARCHAR(255)     DEFAULT NULL,
	created_at      TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY idx_audit_user_id (user_id),
	CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
	CONSTRAINT fk_audit_provider FOREIGN KEY (provider_id) REFERENCES auth_providers(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO auth_providers (code, name) VALUES
	('google', 'Google'),
	('facebook', 'Facebook'),
	('apple', 'Apple'),
	('github', 'GitHub');