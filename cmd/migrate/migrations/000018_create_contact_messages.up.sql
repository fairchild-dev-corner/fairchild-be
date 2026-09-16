CREATE TABLE contact_messages (
	id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	name       VARCHAR(255)  NOT NULL,
	email      VARCHAR(255)  NOT NULL,
	message    VARCHAR(2000) NOT NULL,
	created_at TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY idx_contact_messages_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
