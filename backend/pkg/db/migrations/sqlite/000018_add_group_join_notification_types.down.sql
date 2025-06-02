-- Remove new notification types for group join approval and rejection
-- First, delete any records with the new notification types
DELETE FROM notifications WHERE type IN ('group_join_approved', 'group_join_rejected');

-- Create a new table with the original constraint
CREATE TABLE notifications_new (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('follow_request', 'follow_accepted', 'new_follower', 'post_like', 'post_comment', 'group_invite', 'group_join_request', 'event_invite')),
    content TEXT NOT NULL,
    data TEXT,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from the old table to the new table
INSERT INTO notifications_new (id, user_id, sender_id, type, content, data, read_at, created_at)
SELECT id, user_id, sender_id, type, content, data, read_at, created_at
FROM notifications;

-- Drop the old table
DROP TABLE notifications;

-- Rename the new table to the original name
ALTER TABLE notifications_new RENAME TO notifications;
