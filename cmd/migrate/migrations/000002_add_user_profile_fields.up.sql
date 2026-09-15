ALTER TABLE users
	ADD COLUMN first_name     VARCHAR(100) DEFAULT NULL AFTER username,
	ADD COLUMN last_name      VARCHAR(100) DEFAULT NULL AFTER first_name,
	ADD COLUMN badge_number_id VARCHAR(50) DEFAULT NULL AFTER last_name,
	ADD COLUMN date_of_birth  DATE         DEFAULT NULL AFTER badge_number_id,
	ADD COLUMN gender         VARCHAR(20)  DEFAULT NULL AFTER date_of_birth,
	ADD COLUMN mobile_number  VARCHAR(20)  DEFAULT NULL AFTER gender,
	ADD COLUMN location       VARCHAR(255) DEFAULT NULL AFTER mobile_number,
	ADD UNIQUE KEY uq_users_badge_number_id (badge_number_id);
