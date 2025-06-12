-- SQLite doesn't support modifying CHECK constraints directly
-- We need to recreate the table with the updated constraint

-- Create new posts table with updated visibility constraint
CREATE TABLE posts_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image TEXT,
    visibility TEXT NOT NULL CHECK (visibility IN ('public', 'followers', 'private', 'custom')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old table
INSERT INTO posts_new (id, user_id, content, image, visibility, created_at, updated_at)
SELECT id, user_id, content, image, visibility, created_at, updated_at FROM posts;

-- Drop old table
DROP TABLE posts;

-- Rename new table to original name
ALTER TABLE posts_new RENAME TO posts;
