-- Restore redundant user fields
ALTER TABLE users ADD COLUMN nickname TEXT;
ALTER TABLE users ADD COLUMN about_me TEXT;
