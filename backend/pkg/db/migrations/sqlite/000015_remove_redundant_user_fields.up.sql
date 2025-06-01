-- Remove redundant user fields
-- Remove nickname column (use username instead)
-- Remove about_me column (use bio instead)

-- First, migrate any existing about_me data to bio field
UPDATE users SET bio = about_me WHERE bio IS NULL OR bio = '' AND about_me IS NOT NULL AND about_me != '';

-- Drop the redundant columns
ALTER TABLE users DROP COLUMN nickname;
ALTER TABLE users DROP COLUMN about_me;
