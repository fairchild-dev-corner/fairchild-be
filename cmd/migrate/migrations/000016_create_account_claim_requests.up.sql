CREATE TABLE account_claim_requests (
	id                     BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	member_id              VARCHAR(50)   NOT NULL,
	full_name              VARCHAR(255)  DEFAULT NULL,
	proposed_email         VARCHAR(255)  DEFAULT NULL,
	proposed_mobile_number VARCHAR(20)   DEFAULT NULL,
	message                VARCHAR(1000) DEFAULT NULL,
	status                 ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending',
	rejection_reason       VARCHAR(500)  DEFAULT NULL,
	reviewed_at            TIMESTAMP     NULL DEFAULT NULL,
	created_at             TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at             TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY idx_claim_requests_member_id (member_id),
	KEY idx_claim_requests_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
