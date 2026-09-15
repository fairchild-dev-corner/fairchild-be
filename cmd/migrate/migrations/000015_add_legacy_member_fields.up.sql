ALTER TABLE users
	MODIFY COLUMN status ENUM('active','pending_review','suspended','deleted','inactive') NOT NULL DEFAULT 'active',
	ADD COLUMN nationality     VARCHAR(100) DEFAULT NULL AFTER location,
	ADD COLUMN membership_date DATE         DEFAULT NULL AFTER nationality,
	ADD COLUMN user_client_id  VARCHAR(50)  DEFAULT NULL AFTER membership_date;
