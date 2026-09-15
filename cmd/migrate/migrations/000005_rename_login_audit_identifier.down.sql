ALTER TABLE login_audit_logs
	CHANGE COLUMN identifier_attempted email_attempted VARCHAR(255) DEFAULT NULL,
	DROP KEY idx_audit_identifier_created,
	ADD INDEX idx_audit_email_created (email_attempted, created_at);
