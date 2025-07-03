package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GroupMemberRole represents the role of a group member

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

// PromoteToAdmin promotes a group member to admin
func (s *GroupMemberService) PromoteToAdmin(groupID, memberID, callerID string) error {
	// Check if caller is a group admin
	isAdmin, err := s.IsGroupAdmin(groupID, callerID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return errors.New("only group admins can promote members")
	}

	// Check if the member to be promoted exists and is a regular member
	var memberRole string
	err = s.DB.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, memberID).Scan(&memberRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("member not found in group")
		}
		return fmt.Errorf("failed to get member role: %w", err)
	}

	if memberRole != string(GroupMemberRoleMember) {
		return errors.New("can only promote regular members to admin")
	}

	// Update the member's role to admin
	_, err = s.DB.Exec(`
		UPDATE group_members
		SET role = ?, updated_at = ?
		WHERE group_id = ? AND user_id = ?
	`, GroupMemberRoleAdmin, time.Now(), groupID, memberID)
	if err != nil {
		return fmt.Errorf("failed to promote member to admin: %w", err)
	}

	return nil
}

// DemoteFromAdmin demotes a group admin to a regular member
func (s *GroupMemberService) DemoteFromAdmin(groupID, memberID, callerID string) error {
	// Check if caller is a group admin
	isAdmin, err := s.IsGroupAdmin(groupID, callerID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return errors.New("only group admins can demote members")
	}

	// Prevent self-demotion
	if memberID == callerID {
		return errors.New("cannot demote yourself")
	}

	// Check if the member to be demoted exists and is an admin
	var memberRole string
	err = s.DB.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, memberID).Scan(&memberRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("member not found in group")
		}
		return fmt.Errorf("failed to get member role: %w", err)
	}

	if memberRole != string(GroupMemberRoleAdmin) {
		return errors.New("can only demote admin members")
	}

	// Update the member's role to member
	_, err = s.DB.Exec(`
		UPDATE group_members
		SET role = ?, updated_at = ?
		WHERE group_id = ? AND user_id = ?
	`, GroupMemberRoleMember, time.Now(), groupID, memberID)
	if err != nil {
		return fmt.Errorf("failed to demote admin: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a group based on roles
func (s *GroupMemberService) RemoveMember(groupID, memberID, callerID string) error {
	// Check if caller is a group admin
	isCallerAdmin, err := s.IsGroupAdmin(groupID, callerID)
	if err != nil {
		return fmt.Errorf("failed to check caller admin status: %w", err)
	}
	if !isCallerAdmin {
		return errors.New("only group admins can remove members")
	}

	// Get the member to be removed
	member, err := s.GetByGroupAndUser(groupID, memberID)
	if err != nil {
		return fmt.Errorf("member not found in group: %w", err)
	}

	// Prevent self-removal
	if memberID == callerID {
		return errors.New("cannot remove yourself from the group")
	}

	// Admins can only remove regular members, not other admins
	if member.Role == GroupMemberRoleAdmin || member.Role == GroupMemberRoleCreator {
		return errors.New("admins cannot remove other admins")
	}

	// Delete the member
	_, err = s.DB.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", groupID, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}

	return nil
}

// Delete deletes a group member
// This function is now a wrapper around LeaveGroup for self-removal scenarios.
// For admin-initiated removals, use RemoveMember instead.
func (s *GroupMemberService) Delete(groupID, userID string) error {
	// This is typically used for self-removal (leaving group)
	return s.LeaveGroup(groupID, userID)
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

// GetInvitedUsers retrieves users who have been invited to a group but haven't responded yet
func (s *GroupMemberService) GetInvitedUsers(groupID string) ([]*GroupMember, error) {
	rows, err := s.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'invited'
		ORDER BY gm.created_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invited users: %w", err)
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
			return nil, fmt.Errorf("failed to scan invited user: %w", err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invited users: %w", err)
	}

	return members, nil
}

// IsGroupAdmin checks if a user is an admin of a group
func (s *GroupMemberService) IsGroupAdmin(groupID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(`
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND (role = 'admin' OR role = 'creator') AND status = 'accepted'
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

// Update updates a group member
func (s *GroupMemberService) Update(member *GroupMember) error {
	member.UpdatedAt = time.Now()

	_, err := s.DB.Exec(`
		UPDATE group_members
		SET role = ?, status = ?, updated_at = ?
		WHERE id = ?
	`, member.Role, member.Status, member.UpdatedAt, member.ID)
	if err != nil {
		return fmt.Errorf("failed to update group member: %w", err)
	}

	return nil
}

// GetGroupAdmins retrieves all admins of a group (including creator)
func (s *GroupMemberService) GetGroupAdmins(groupID string) ([]*GroupMember, error) {
	// First get all admin members
	rows, err := s.DB.Query(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.role = 'admin' AND gm.status = 'accepted'
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group admins: %w", err)
	}
	defer rows.Close()

	var admins []*GroupMember
	adminUserIDs := make(map[string]bool) // Track unique user IDs

	// Add all admin members
	for rows.Next() {
		admin := &GroupMember{User: &User{}}
		err := rows.Scan(
			&admin.ID, &admin.GroupID, &admin.UserID, &admin.Role, &admin.Status, &admin.CreatedAt, &admin.UpdatedAt,
			&admin.User.ID, &admin.User.Username, &admin.User.FullName, &admin.User.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group admin: %w", err)
		}

		admins = append(admins, admin)
		adminUserIDs[admin.UserID] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group admins: %w", err)
	}

	// Now get the group creator if they're not already included as an admin
	var creatorID string
	err = s.DB.QueryRow(`
		SELECT g.creator_id
		FROM groups g
		WHERE g.id = ?
	`, groupID).Scan(&creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group creator: %w", err)
	}

	// If creator is not already in the admins list, add them
	if !adminUserIDs[creatorID] {
		var creator GroupMember
		creator.User = &User{}
		err = s.DB.QueryRow(`
			SELECT u.id, u.username, u.full_name, u.profile_picture
			FROM users u
			WHERE u.id = ?
		`, creatorID).Scan(
			&creator.User.ID, &creator.User.Username, &creator.User.FullName, &creator.User.ProfilePicture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get creator user info: %w", err)
		}

		creator.GroupID = groupID
		creator.UserID = creatorID
		creator.Role = GroupMemberRoleAdmin
		creator.Status = GroupMemberStatusAccepted
		admins = append(admins, &creator)
	}

	return admins, nil
}

// GetMemberWithMostPosts finds the group member with the most posts
func (s *GroupMemberService) GetMemberWithMostPosts(groupID string) (*GroupMember, error) {
	member := &GroupMember{User: &User{}}

	err := s.DB.QueryRow(`
		SELECT gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture,
			COUNT(gp.id) as post_count
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		LEFT JOIN group_posts gp ON gm.user_id = gp.user_id AND gm.group_id = gp.group_id
		WHERE gm.group_id = ? AND gm.status = 'accepted' AND gm.role = 'member'
		GROUP BY gm.id, gm.group_id, gm.user_id, gm.role, gm.status, gm.created_at, gm.updated_at,
			u.id, u.username, u.full_name, u.profile_picture
		ORDER BY post_count DESC, gm.created_at ASC
		LIMIT 1
	`, groupID).Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Role, &member.Status, &member.CreatedAt, &member.UpdatedAt,
		&member.User.ID, &member.User.Username, &member.User.FullName, &member.User.ProfilePicture,
		new(int), // post_count - we don't need to store this
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no eligible members found")
		}
		return nil, fmt.Errorf("failed to get member with most posts: %w", err)
	}

	return member, nil
}

// TransferOwnership handles transferring group ownership when the creator leaves
func (s *GroupMemberService) TransferOwnership(groupID, leavingCreatorID string) error {
	// Start a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove the creator from group_members table
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", groupID, leavingCreatorID)
	if err != nil {
		return fmt.Errorf("failed to remove creator from group members: %w", err)
	}

	// Check if there are any existing admins
	var adminCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM group_members WHERE group_id = ? AND role = 'admin' AND status = 'accepted'", groupID).Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to count admins: %w", err)
	}

	// If no admins exist, promote the member with the most posts
	if adminCount == 0 {
		// Find member with most posts using a transaction-aware query
		var newAdminID string
		err = tx.QueryRow(`
			SELECT gm.user_id
			FROM group_members gm
			LEFT JOIN group_posts gp ON gm.user_id = gp.user_id AND gm.group_id = gp.group_id
			WHERE gm.group_id = ? AND gm.status = 'accepted' AND gm.role = 'member'
			GROUP BY gm.user_id
			ORDER BY COUNT(gp.id) DESC, gm.created_at ASC
			LIMIT 1
		`, groupID).Scan(&newAdminID)

		if err != nil {
			if err == sql.ErrNoRows {
				// No members left, group becomes ownerless but remains active
				return tx.Commit()
			}
			return fmt.Errorf("failed to find member to promote: %w", err)
		}

		// Promote the selected member to admin
		_, err = tx.Exec(`
			UPDATE group_members
			SET role = ?, updated_at = ?
			WHERE group_id = ? AND user_id = ?
		`, GroupMemberRoleAdmin, time.Now(), groupID, newAdminID)
		if err != nil {
			return fmt.Errorf("failed to promote member to admin: %w", err)
		}
	}

	// Commit the transaction
	return tx.Commit()
}

// LeaveGroup handles a user leaving a group with proper admin succession logic
func (s *GroupMemberService) LeaveGroup(groupID, userID string) error {
	// Start a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the user's current role in the group
	var userRole string
	err = tx.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`, groupID, userID).Scan(&userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user is not a member of this group")
		}
		return fmt.Errorf("failed to get user role: %w", err)
	}

	// Remove the user from the group
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ? AND user_id = ?", groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// If the leaving user was an admin or creator, check if we need to assign a new admin
	if userRole == string(GroupMemberRoleAdmin) || userRole == string(GroupMemberRoleCreator) {
		// Count remaining admins
		var adminCount int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM group_members
			WHERE group_id = ? AND (role = 'admin' OR role = 'creator') AND status = 'accepted'
		`, groupID).Scan(&adminCount)
		if err != nil {
			return fmt.Errorf("failed to count remaining admins: %w", err)
		}

		// If no admins remain, promote the earliest member
		if adminCount == 0 {
			var newAdminID string
			err = tx.QueryRow(`
				SELECT user_id FROM group_members
				WHERE group_id = ? AND status = 'accepted' AND role = 'member'
				ORDER BY created_at ASC
				LIMIT 1
			`, groupID).Scan(&newAdminID)

			if err != nil {
				if err == sql.ErrNoRows {
					// No members left, group becomes ownerless but remains active
					return tx.Commit()
				}
				return fmt.Errorf("failed to find member to promote: %w", err)
			}

			// Promote the earliest member to admin
			_, err = tx.Exec(`
				UPDATE group_members
				SET role = ?, updated_at = ?
				WHERE group_id = ? AND user_id = ?
			`, GroupMemberRoleAdmin, time.Now(), groupID, newAdminID)
			if err != nil {
				return fmt.Errorf("failed to promote member to admin: %w", err)
			}
		}
	}

	// Commit the transaction
	return tx.Commit()
}
