-- Add status column to notifications table
ALTER TABLE notifications ADD COLUMN status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined', 'approved', 'rejected'));
