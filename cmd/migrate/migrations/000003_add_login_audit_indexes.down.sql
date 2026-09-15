ALTER TABLE login_audit_logs
	DROP KEY idx_audit_email_created,
	DROP KEY idx_audit_ip_created;
