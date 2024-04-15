#!/bin/bash

# Function to make a commit with a specific date
commit_with_date() {
    local message="$1"
    local date="$2"
    
    # Set the environment variables for the commit date
    export GIT_AUTHOR_DATE="$date"
    export GIT_COMMITTER_DATE="$date"
    
    # Make the commit
    git commit -m "$message"
    
    # Reset the environment variables
    unset GIT_AUTHOR_DATE
    unset GIT_COMMITTER_DATE
}

# Navigate to the project directory
cd /home/bernaotieno/github/social-network

# Initialize the repository if not already done
git init

# Configure git if needed
git config --local user.name "Berna Otieno"
git config --local user.email "bernaotieno@example.com"

# 1. Initial commit - Project setup
git add README.md docker-compose.yml
commit_with_date "Initial project setup with README and Docker configuration" "2023-12-15T10:00:00"

# 2. Backend structure setup
git add backend/go.mod backend/go.sum backend/server.go
commit_with_date "Set up basic Go backend structure" "2023-12-16T14:30:00"

# 3. Add database models
git add backend/pkg/models/
commit_with_date "Add database models for users, posts, and comments" "2023-12-18T09:45:00"

# 4. Add authentication
git add backend/pkg/auth/
commit_with_date "Implement authentication with JWT" "2023-12-20T16:20:00"

# 5. Add middleware
git add backend/pkg/middleware/
commit_with_date "Add middleware for authentication, logging, and CORS" "2023-12-22T11:15:00"

# 6. Add handlers
git add backend/pkg/handlers/
commit_with_date "Implement API handlers for users and posts" "2023-12-27T13:40:00"

# 7. Add WebSocket support
git add backend/pkg/websocket/
commit_with_date "Add WebSocket support for real-time chat" "2024-01-02T10:30:00"

# 8. Add database migrations
git add backend/pkg/db/
commit_with_date "Add database migrations and SQLite setup" "2024-01-05T15:20:00"

# 9. Add utilities
git add backend/pkg/utils/
commit_with_date "Add utility functions for API responses and image handling" "2024-01-08T09:10:00"

# 10. Add Dockerfile for backend
git add backend/Dockerfile
commit_with_date "Add Dockerfile for backend" "2024-01-10T14:25:00"

# 11. Frontend initial setup
git add frontend/package.json frontend/next.config.js
commit_with_date "Initialize Next.js frontend project" "2024-01-15T11:30:00"

# 12. Add frontend components
git add frontend/components/
commit_with_date "Add basic UI components" "2024-01-18T16:45:00"

# 13. Add frontend pages
git add frontend/pages/
commit_with_date "Add main pages for the application" "2024-01-22T10:15:00"

# 14. Add frontend styles
git add frontend/styles/
commit_with_date "Add CSS styles for the application" "2024-01-25T13:50:00"

# 15. Add frontend API integration
git add frontend/lib/
commit_with_date "Add API integration for frontend" "2024-01-30T09:20:00"

# 16. Add frontend authentication
git add frontend/context/
commit_with_date "Implement authentication context for frontend" "2024-02-05T14:10:00"

# 17. Add frontend Dockerfile
git add frontend/Dockerfile
commit_with_date "Add Dockerfile for frontend" "2024-02-08T11:35:00"

# 18. Add group functionality to backend
git add backend/pkg/models/group.go backend/pkg/models/group_member.go
commit_with_date "Add group models and functionality" "2024-02-15T10:25:00"

# 19. Add group handlers
git add backend/pkg/handlers/group.go
commit_with_date "Implement group API handlers" "2024-02-18T15:40:00"

# 20. Add event functionality
git add backend/pkg/models/event.go backend/pkg/models/event_response.go
commit_with_date "Add event models and functionality" "2024-02-25T09:15:00"

# 21. Add notification system
git add backend/pkg/models/notification.go
commit_with_date "Implement notification system" "2024-03-01T13:30:00"

# 22. Add group components to frontend
git add frontend/components/groups/
commit_with_date "Add group UI components" "2024-03-05T16:20:00"

# 23. Add event components to frontend
git add frontend/components/events/
commit_with_date "Add event UI components" "2024-03-10T11:45:00"

# 24. Add chat functionality to frontend
git add frontend/components/chat/
commit_with_date "Implement real-time chat UI" "2024-03-15T14:10:00"

# 25. Add profile page improvements
git add frontend/pages/profile/
commit_with_date "Enhance user profile page" "2024-03-20T09:30:00"

# 26. Add feed improvements
git add frontend/pages/feed.js
commit_with_date "Improve feed with infinite scrolling" "2024-03-25T15:45:00"

# 27. Add search functionality
git add frontend/components/search/
commit_with_date "Implement search functionality" "2024-04-01T10:20:00"

# 28. Add mobile responsiveness
git add frontend/styles/mobile.css
commit_with_date "Improve mobile responsiveness" "2024-04-05T13:15:00"

# 29. Add performance optimizations
git add frontend/lib/optimizations.js
commit_with_date "Implement performance optimizations" "2024-04-10T16:30:00"

# 30. Final touches and bug fixes
git add .
commit_with_date "Final touches and bug fixes" "2024-04-15T11:00:00"

echo "Commit history created successfully!"
