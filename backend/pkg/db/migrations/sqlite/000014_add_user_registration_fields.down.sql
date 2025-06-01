-- Remove the new registration fields
ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
ALTER TABLE users DROP COLUMN date_of_birth;
ALTER TABLE users DROP COLUMN nickname;
ALTER TABLE users DROP COLUMN about_me;
