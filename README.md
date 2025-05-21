# Social Network

A Facebook-like social network application with a Next.js frontend and Go backend.

## Features

- 🔐 **Authentication**: User registration, login, and session management
- 👤 **Profiles**: User profiles with customizable privacy settings
- 🧑‍🤝‍🧑 **Follow System**: Follow/unfollow users with privacy controls
- 📝 **Posts & Comments**: Create, read, update, and delete posts and comments
- 💬 **Real-time Chat**: Private messaging and group chats via WebSockets
- 👥 **Groups**: Create and join groups, share posts within groups
- 🔔 **Notifications**: Real-time notifications for various activities

## Tech Stack

### Frontend
- **Framework**: Next.js (JavaScript)
- **State Management**: React Context API
- **Styling**: CSS Modules
- **Real-time Communication**: WebSockets

### Backend
- **Language**: Go
- **Database**: SQLite
- **Authentication**: Session-based with cookies
- **Real-time Communication**: WebSockets (gorilla/websocket)
- **Password Hashing**: bcrypt
- **File Storage**: Local file system

### Infrastructure
- **Containerization**: Docker
- **Database Migrations**: golang-migrate

## Project Structure

```
/
├── frontend/               # Next.js frontend
│   ├── src/
│   │   ├── app/            # Next.js App Router
│   │   ├── components/     # Reusable components
│   │   ├── hooks/          # Custom React hooks
│   │   ├── styles/         # CSS modules
│   │   └── utils/          # Utility functions
│   └── public/             # Static assets
│
├── backend/                # Go backend
│   ├── pkg/
│   │   ├── auth/           # Authentication logic
│   │   ├── db/             # Database connection and migrations
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── models/         # Database models
│   │   ├── utils/          # Utility functions
│   │   └── websocket/      # WebSocket implementation
│   └── server.go           # Main server file
│
├── uploads/                # Uploaded files
│   ├── avatars/            # User profile pictures
│   └── posts/              # Post images
│
└── docker-compose.yml      # Docker Compose configuration
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Node.js (for local development)
- Go (for local development)

### Running with Docker

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/social-network.git
   cd social-network
   ```

2. Start the application with Docker Compose:
   ```bash
   docker-compose up
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080/api

### Local Development

#### Frontend

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Access the frontend at http://localhost:3000

#### Backend

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the server:
   ```bash
   go run server.go
   ```

4. Access the API at http://localhost:8080/api

## API Documentation

### Authentication

- `POST /api/auth/register`: Register a new user
- `POST /api/auth/login`: Login a user
- `POST /api/auth/logout`: Logout a user

### Users

- `GET /api/users`: Get a list of users
- `GET /api/users/{id}`: Get a user by ID
- `PUT /api/users/profile`: Update user profile
- `POST /api/users/avatar`: Upload profile picture
- `POST /api/users/{id}/follow`: Follow a user
- `DELETE /api/users/{id}/follow`: Unfollow a user
- `GET /api/users/{id}/followers`: Get user's followers
- `GET /api/users/{id}/following`: Get users followed by user
- `GET /api/users/follow-requests`: Get pending follow requests
- `PUT /api/users/follow-requests/{id}`: Respond to a follow request

### Posts

- `POST /api/posts`: Create a new post
- `GET /api/posts/feed`: Get posts for user's feed
- `GET /api/posts/user/{id}`: Get posts by a user
- `GET /api/posts/{id}`: Get a post by ID
- `PUT /api/posts/{id}`: Update a post
- `DELETE /api/posts/{id}`: Delete a post
- `POST /api/posts/{id}/like`: Like a post
- `DELETE /api/posts/{id}/like`: Unlike a post
- `GET /api/posts/{id}/comments`: Get comments for a post
- `POST /api/posts/{id}/comments`: Add a comment to a post
- `DELETE /api/posts/{postId}/comments/{commentId}`: Delete a comment

### Groups

- `GET /api/groups`: Get a list of groups
- `POST /api/groups`: Create a new group
- `GET /api/groups/{id}`: Get a group by ID
- `PUT /api/groups/{id}`: Update a group
- `DELETE /api/groups/{id}`: Delete a group
- `POST /api/groups/{id}/join`: Join a group
- `DELETE /api/groups/{id}/join`: Leave a group
- `GET /api/groups/{id}/posts`: Get posts in a group
- `POST /api/groups/{id}/posts`: Create a post in a group
- `GET /api/groups/{id}/events`: Get events in a group
- `POST /api/groups/{id}/events`: Create an event in a group
- `POST /api/groups/events/{id}/respond`: Respond to an event

### Notifications

- `GET /api/notifications`: Get user's notifications
- `PUT /api/notifications/{id}/read`: Mark a notification as read
- `PUT /api/notifications/read-all`: Mark all notifications as read

### WebSocket

- `/ws`: WebSocket endpoint for real-time communication

## License

This project is licensed under the MIT License - see the LICENSE file for details.
