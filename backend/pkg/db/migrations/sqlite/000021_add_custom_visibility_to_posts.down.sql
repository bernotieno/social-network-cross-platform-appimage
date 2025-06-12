-- Revert the posts table to the original constraint (remove 'custom' visibility)

-- Create posts table with original constraint
CREATE TABLE posts_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image TEXT,
    visibility TEXT NOT NULL CHECK (visibility IN ('public', 'followers', 'private')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old table (excluding any posts with 'custom' visibility)
INSERT INTO posts_new (id, user_id, content, image, visibility, created_at, updated_at)
SELECT id, user_id, content, image, visibility, created_at, updated_at 
FROM posts 
WHERE visibility IN ('public', 'followers', 'private');

-- Drop old table
DROP TABLE posts;

-- Rename new table to original name
ALTER TABLE posts_new RENAME TO posts;
