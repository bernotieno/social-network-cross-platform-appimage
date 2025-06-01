-- Remove foreign key constraints from likes and comments tables to allow group posts
-- SQLite doesn't support dropping foreign keys directly, so we need to recreate the tables

-- Create new likes table without foreign key constraint to posts
CREATE TABLE likes_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (post_id, user_id)
);

-- Copy data from old table
INSERT INTO likes_new SELECT * FROM likes;

-- Drop old table and rename new one
DROP TABLE likes;
ALTER TABLE likes_new RENAME TO likes;

-- Create new comments table without foreign key constraint to posts
CREATE TABLE comments_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old table
INSERT INTO comments_new SELECT * FROM comments;

-- Drop old table and rename new one
DROP TABLE comments;
ALTER TABLE comments_new RENAME TO comments;
