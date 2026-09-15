ALTER TABLE users
	DROP COLUMN user_client_id,
	DROP COLUMN membership_date,
	DROP COLUMN nationality,
	MODIFY COLUMN status ENUM('active','pending_review','suspended','deleted') NOT NULL DEFAULT 'active';
