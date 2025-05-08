
# ✅ Facebook-like Social Network Development Plan

## 📁 Phase 1: Planning & Architecture (Day 1–2)
- [ ] Define core features (posts, followers, groups, chats, notifications)
- [ ] Design database schema (ER Diagram)
- [ ] Choose tech stack (Go, SQLite, JS framework, Docker)
- [ ] Initialize Go project structure
- [ ] Set up Git repository
- [ ] Write `Dockerfile` and `docker-compose.yml`
- [ ] Add README with setup instructions

## 🔐 Phase 2: Authentication & Profiles (Day 3–5)
- [ ] Create `User` model and migrations
- [ ] Implement registration and login endpoints
- [ ] Implement session/cookie or JWT-based auth
- [ ] Protect routes with middleware
- [ ] Create user profile view/edit endpoints
- [ ] Support profile picture and bio updates
- [ ] Build basic UI for login/register/profile

## 🤝 Phase 3: Social Graph – Follow System (Day 6–8)
- [ ] Create `FollowRequest` table (pending, accepted, rejected)
- [ ] Endpoint: send follow request
- [ ] Endpoint: cancel/accept/reject request
- [ ] Display follower/following count on profiles
- [ ] Enforce privacy for profiles and posts
- [ ] UI for follow/unfollow and request status

## 📝 Phase 4: Posts & Feed (Day 9–13)
- [ ] Create `Post` model (content, image, visibility)
- [ ] Endpoint: create/edit/delete post
- [ ] Endpoint: get user’s own posts
- [ ] Endpoint: feed (followed users + public posts)
- [ ] Add pagination to feed/posts
- [ ] Build feed and post creation UI

## 👥 Phase 5: Groups & Events (Day 14–18)
- [ ] Create `Group` model (name, description, privacy)
- [ ] Join/leave group flow with approvals
- [ ] Create `GroupPost` and `GroupEvent` models
- [ ] Endpoints for posting in groups and creating events
- [ ] UI: group listing, detail page, group post/event creation

## 💬 Phase 6: Messaging (Day 19–23)
- [ ] Setup WebSocket server (gorilla/websocket)
- [ ] Handle user connection/authentication
- [ ] Create `Message` model (sender, receiver, timestamp)
- [ ] Store and retrieve direct + group messages
- [ ] Frontend real-time chat interface
- [ ] Message delivery indicators (sent/read)

## 🔔 Phase 7: Notifications (Day 24–26)
- [ ] Create `Notification` model (type, content, read_at)
- [ ] Hook notifications into:
  - [ ] Follow requests
  - [ ] Group invites
  - [ ] New messages
- [ ] API to list + mark notifications as read
- [ ] Push notifications via WebSocket or polling
- [ ] UI notification panel

## 🚀 Phase 8: Finalization, Testing & Deployment (Day 27–30)
- [ ] Write unit tests for core API endpoints
- [ ] Load test WebSocket server and REST API
- [ ] Responsive UI testing
- [ ] CI/CD setup (GitHub Actions or other)
- [ ] Final Docker build and deployment
- [ ] Publish final README, API docs, and .env.example
