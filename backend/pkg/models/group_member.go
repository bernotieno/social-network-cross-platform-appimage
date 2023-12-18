package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GroupMemberRole represents the role of a group member
type GroupMemberRole string

const (
	GroupMemberRoleAdmin  GroupMemberRole = "admin"
	GroupMemberRoleMember GroupMemberRole = "member"
)

// GroupMemberStatus represents the status of a group membership
type GroupMemberStatus string

const (
	GroupMemberStatusPending  GroupMemberStatus = "pending"
	GroupMemberStatusAccepted GroupMemberStatus = "accepted"
	GroupMemberStatusRejected GroupMemberStatus = "rejected"
)

// GroupMember represents a member of a group
type GroupMember struct {
	ID        string            `json:"id"`
	GroupID   string            `json:"groupId"`
	UserID    string            `json:"userId"`
	Role      GroupMemberRole   `json:"role"`
	Status    GroupMemberStatus `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
	// Additional fields for API responses
	User  *User  `json:"user,omitempty"`
	Group *Group `json:"group,omitempty"`
}

// GroupMemberService handles group member-related operations
type GroupMemberService struct {
	DB *sql.DB
}

// NewGroupMemberService creates a new GroupMemberService
func NewGroupMemberService(db *sql.DB) *GroupMemberService {
	return &GroupMemberService{DB: db}
}

// Create creates a new group member
func (s *GroupMemberService) Create(member *GroupMember) error {
	member.ID = uuid.New().String()
	now := time.Now()
	member.CreatedAt = now
	member.UpdatedAt = now

	_, err := s.DB.Exec(`
		INSERT INTO group_members (id, group_id, user_id, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, member.ID, member.GroupID, member.UserID, member.Role, member.Status, member.CreatedAt, member.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create group member: %w", err)
	}

	return nil
}

// GetByID retrieves a group member by ID
func (s *GroupMemberService) GetByID(id string) (*GroupMember, error) {
	member := &GroupMember{User: &User{}, Group: &Group{}}
	err := s.DB.QueryRow(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture,
			g.id, g.name, g.privacy
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		JOIN groups g ON gm.group_id = g.id
		WHERE gm.id = ?
	`, id).Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Role, &member.Status, &member.CreatedAt, &member.UpdatedAt,
		&member.User.ID, &member.User.Username, &member.User.FullName, &member.User.ProfilePicture,
		&member.Group.ID, &member.Group.Name, &member.Group.Privacy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("group member not found")
		}
		return nil, fmt.Errorf("failed to get group member: %w", err)
	}

	return member, nil
}

// GetByGroupAndUser retrieves a group member by group ID and user ID
func (s *GroupMemberService) GetByGroupAndUser(groupID, userID string) (*GroupMember, error) {
	member := &GroupMember{}
	err := s.DB.QueryRow(`
		SELECT id, group_id, user_id, role, status, created_at, updated_at
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Role, &member.Status, &member.CreatedAt, &member.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("group member not found")
		}
		return nil, fmt.Errorf("failed to get group member: %w", err)
	}

	return member, nil
}

// UpdateStatus updates the status of a group member
func (s *GroupMemberService) UpdateStatus(id string, status GroupMemberStatus) error {
	_, err := s.DB.Exec(`
		UPDATE group_members
		SET status = ?, updated_at = ?
		WHERE id = ?
	`, status, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to update group member status: %w", err)
	}

	return nil
}

// UpdateRole updates the role of a group member
func (s *GroupMemberService) UpdateRole(id string, role GroupMemberRole) error {
	_, err := s.DB.Exec(`
		UPDATE group_members
		SET role = ?, updated_at = ?
		WHERE id = ?
	`, role, time.Now(), id)

	if err != nil {
		return fmt.Errorf("failed to update group member role: %w", err)
	}

	return nil
}

// Delete deletes a group member
func (s *GroupMemberService) Delete(groupID, userID string) error {
	_, err := s.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete group member: %w", err)
	}

	return nil
}

// GetMembers retrieves members of a group
func (s *GroupMemberService) GetMembers(groupID string, limit, offset int) ([]*GroupMember, error) {
	rows, err := s.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'accepted'
		ORDER BY gm.role = 'admin' DESC, gm.created_at ASC
		LIMIT ? OFFSET ?
	`, groupID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []*GroupMember
	for rows.Next() {
		member := &GroupMember{User: &User{}}
		err := rows.Scan(
			&member.ID, &member.GroupID, &member.UserID, &member.Role, &member.Status, &member.CreatedAt, &member.UpdatedAt,
			&member.User.ID, &member.User.Username, &member.User.FullName, &member.User.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group members: %w", err)
	}

	return members, nil
}

// GetPendingRequests retrieves pending join requests for a group
func (s *GroupMemberService) GetPendingRequests(groupID string) ([]*GroupMember, error) {
	rows, err := s.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'pending'
		ORDER BY gm.created_at ASC
	`, groupID)

	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	defer rows.Close()

	var members []*GroupMember
	for rows.Next() {
		member := &GroupMember{User: &User{}}
		err := rows.Scan(
			&member.ID, &member.GroupID, &member.UserID, &member.Role, &member.Status, &member.CreatedAt, &member.UpdatedAt,
			&member.User.ID, &member.User.Username, &member.User.FullName, &member.User.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending requests: %w", err)
	}

	return members, nil
}

// IsGroupAdmin checks if a user is an admin of a group
func (s *GroupMemberService) IsGroupAdmin(groupID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND role = 'admin' AND status = 'accepted'
	`, groupID, userID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check if user is admin: %w", err)
	}

	return count > 0, nil
}

// IsGroupMember checks if a user is a member of a group
func (s *GroupMemberService) IsGroupMember(groupID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, userID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check if user is member: %w", err)
	}

	return count > 0, nil
}
