ALTER TABLE user_account_cred
	ADD COLUMN password_change_required TINYINT(1) NOT NULL DEFAULT 0 AFTER password_hash;
