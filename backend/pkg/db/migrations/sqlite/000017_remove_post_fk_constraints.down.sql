-- Restore foreign key constraints to likes and comments tables

-- Create likes table with foreign key constraint to posts
CREATE TABLE likes_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (post_id, user_id)
);

-- Copy data from old table (only posts that exist in posts table)
INSERT INTO likes_new 
SELECT l.* FROM likes l 
WHERE EXISTS (SELECT 1 FROM posts p WHERE p.id = l.post_id);

-- Drop old table and rename new one
DROP TABLE likes;
ALTER TABLE likes_new RENAME TO likes;

-- Create comments table with foreign key constraint to posts
CREATE TABLE comments_new (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from old table (only posts that exist in posts table)
INSERT INTO comments_new 
SELECT c.* FROM comments c 
WHERE EXISTS (SELECT 1 FROM posts p WHERE p.id = c.post_id);

-- Drop old table and rename new one
DROP TABLE comments;
ALTER TABLE comments_new RENAME TO comments;
