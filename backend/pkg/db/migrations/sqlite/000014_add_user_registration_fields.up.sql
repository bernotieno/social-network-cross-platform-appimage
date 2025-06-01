-- Add new fields for enhanced user registration
ALTER TABLE users ADD COLUMN first_name TEXT;
ALTER TABLE users ADD COLUMN last_name TEXT;
ALTER TABLE users ADD COLUMN date_of_birth DATE;
ALTER TABLE users ADD COLUMN nickname TEXT;
ALTER TABLE users ADD COLUMN about_me TEXT;

-- Update existing users to split full_name into first_name and last_name
-- This is a simple split on the first space - in production you might want more sophisticated logic
UPDATE users SET 
    first_name = CASE 
        WHEN instr(full_name, ' ') > 0 THEN substr(full_name, 1, instr(full_name, ' ') - 1)
        ELSE full_name
    END,
    last_name = CASE 
        WHEN instr(full_name, ' ') > 0 THEN substr(full_name, instr(full_name, ' ') + 1)
        ELSE ''
    END
WHERE first_name IS NULL;
