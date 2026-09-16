ALTER TABLE users
	ADD COLUMN address VARCHAR(255) DEFAULT NULL AFTER location;

CREATE TABLE user_notification_preferences (
	user_id                BIGINT UNSIGNED NOT NULL,
	notifications_enabled  TINYINT(1) NOT NULL DEFAULT 1,
	created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (user_id),
	CONSTRAINT fk_user_notification_preferences_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
