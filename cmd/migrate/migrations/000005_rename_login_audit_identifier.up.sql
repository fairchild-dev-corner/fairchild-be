ALTER TABLE login_audit_logs
	CHANGE COLUMN email_attempted identifier_attempted VARCHAR(255) DEFAULT NULL,
	DROP KEY idx_audit_email_created,
	ADD INDEX idx_audit_identifier_created (identifier_attempted, created_at);
