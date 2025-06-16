-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE comments_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old table (excluding image column)
INSERT INTO comments_new (id, post_id, user_id, content, created_at, updated_at)
SELECT id, post_id, user_id, content, created_at, updated_at FROM comments;

-- Drop old table and rename new one
DROP TABLE comments;
ALTER TABLE comments_new RENAME TO comments;
