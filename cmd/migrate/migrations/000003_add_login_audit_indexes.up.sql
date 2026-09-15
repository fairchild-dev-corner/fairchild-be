ALTER TABLE login_audit_logs
	ADD INDEX idx_audit_email_created (email_attempted, created_at),
	ADD INDEX idx_audit_ip_created (ip_address, created_at);
