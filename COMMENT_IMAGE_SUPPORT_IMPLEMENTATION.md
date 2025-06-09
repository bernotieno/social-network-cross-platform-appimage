# Comment Image Support Implementation

This document outlines the comprehensive changes made to add image and GIF support to comments in the social network platform.

## Database Changes

### 1. Migration Files Created
- **File**: `backend/pkg/db/migrations/sqlite/000019_add_image_to_comments.up.sql`
- **Content**: `ALTER TABLE comments ADD COLUMN image TEXT;`
- **Purpose**: Adds an optional image field to store image paths for comments

### 2. Migration Rollback
- **File**: `backend/pkg/db/migrations/sqlite/000019_add_image_to_comments.down.sql`
- **Purpose**: Provides rollback functionality to remove the image column

## Backend Changes

### 1. Comment Model Updates (`backend/pkg/models/comment.go`)
- **Added Field**: `Image string` to the Comment struct
- **Updated Methods**:
  - `Create()`: Now includes image field in INSERT statement
  - `GetByID()`: Now selects and scans image field
  - `Update()`: Now updates image field
  - `GetCommentsByPost()`: Now includes image field in queries

### 2. Comment Handler Updates (`backend/pkg/handlers/comment.go`)
- **Enhanced AddComment Handler**:
  - Supports both JSON (text-only) and multipart form data (with images)
  - Validates file types and sizes
  - Saves uploaded images using `utils.SaveImage()`
  - Allows empty content if image is provided

### 3. Group Comment Handler Updates (`backend/pkg/handlers/group.go`)
- **Enhanced AddGroupPostComment Handler**:
  - Same multipart form support as regular comments
  - Maintains group membership validation
  - Supports image uploads for group post comments

## Frontend Changes

### 1. Post Component Updates (`frontend/src/components/Post.js`)

#### New State Variables
- `commentImage`: Stores selected image file
- `commentImagePreview`: Stores preview URL for selected image

#### New Functions
- `handleCommentImageChange()`: Handles image file selection with validation
- `handleRemoveCommentImage()`: Removes selected image
- Enhanced `handleCommentSubmit()`: Supports both text and image submissions

#### UI Enhancements
- **Image Upload Button**: Camera icon in comment form
- **Image Preview**: Shows selected image with remove button
- **Comment Display**: Renders images in comments using appropriate tags (img for GIFs, Image for static)

### 2. API Updates (`frontend/src/utils/api.js`)

#### New Methods
- `addCommentWithImage()`: Handles multipart form data for post comments
- Enhanced `addGroupPostComment()`: Detects FormData and handles accordingly

### 3. CSS Styling (`frontend/src/styles/Post.module.css`)

#### New Styles
- `.commentInputContainer`: Flexbox layout for input and actions
- `.commentActions`: Container for upload button and submit button
- `.commentImageUpload`: Styled camera button for image uploads
- `.commentImagePreview`: Container for image preview
- `.commentPreviewImage`: Styling for preview images
- `.removeCommentImage`: Remove button for preview images
- `.commentImageContainer`: Container for images in comments
- `.commentImage`: Styling for images in comment display

## Key Features Implemented

### ✅ Image Upload in Comments
- **File Types**: JPEG, PNG, GIF supported
- **File Size**: 5MB maximum limit
- **Validation**: Client-side and server-side validation
- **Storage**: Images saved in `uploads/comments/` directory

### ✅ GIF Support in Comments
- **Animation Preserved**: GIFs use regular `<img>` tag to maintain animation
- **Detection**: Automatic GIF detection using `isGif()` utility
- **Fallback**: Static images use optimized Next.js Image component

### ✅ Enhanced UX
- **Preview**: Real-time image preview before submission
- **Validation**: User-friendly error messages for invalid files
- **Flexible Content**: Comments can have text, image, or both
- **Remove Option**: Easy removal of selected images

### ✅ Group Comments Support
- **Consistency**: Same image support for group post comments
- **Permissions**: Maintains group membership validation
- **Real-time**: WebSocket broadcasting for image comments

## File Validation

### Client-Side
- **File Types**: `image/jpeg`, `image/jpg`, `image/png`, `image/gif`
- **File Size**: Maximum 5MB
- **Error Handling**: Alert messages for invalid files

### Server-Side
- **MIME Type**: Validates actual file content
- **File Extension**: Ensures proper extension
- **Size Limit**: 5MB maximum enforced
- **Path Sanitization**: Secure file path generation

## API Endpoints

### Regular Post Comments
- **POST** `/api/posts/{id}/comments`
  - **JSON**: `{"content": "text"}` for text-only
  - **FormData**: `content` + `image` file for image comments

### Group Post Comments
- **POST** `/api/groups/{groupId}/posts/{postId}/comments`
  - **JSON**: `{"content": "text"}` for text-only
  - **FormData**: `content` + `image` file for image comments

## Database Schema

### Updated Comments Table
```sql
CREATE TABLE comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT NOT NULL,
    image TEXT,  -- NEW FIELD
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Testing Instructions

### Manual Testing Steps

1. **Apply Database Migration**:
   ```bash
   cd backend
   sqlite3 social_network.db "ALTER TABLE comments ADD COLUMN image TEXT;"
   ```

2. **Start Application**:
   ```bash
   docker-compose up
   # OR
   cd backend && go run server.go
   cd frontend && npm run dev
   ```

3. **Test Comment Images**:
   - Create a post
   - Add text-only comment
   - Add image-only comment
   - Add comment with both text and image
   - Test with different image formats (JPEG, PNG, GIF)
   - Verify GIF animations work

4. **Test Group Comments**:
   - Join a group
   - Create group post
   - Add image comments to group posts
   - Verify permissions work correctly

5. **Test Validation**:
   - Try uploading non-image files (should fail)
   - Try uploading files over 5MB (should fail)
   - Verify error messages appear

### Expected Results
- ✅ Comments display images inline with text
- ✅ GIFs animate properly in comments
- ✅ Image previews work in comment forms
- ✅ File validation prevents invalid uploads
- ✅ Group comments support images
- ✅ Real-time updates work with image comments

## Performance Considerations

### Optimizations
- **Conditional Rendering**: GIFs use `<img>`, static images use `<Image>`
- **File Size Limits**: Prevent large file uploads
- **Lazy Loading**: Images load as needed
- **Compression**: Server-side image optimization (if implemented)

### Future Improvements
- **Thumbnail Generation**: Create thumbnails for large images
- **Image Compression**: Client-side compression before upload
- **CDN Integration**: Store images on CDN for better performance
- **Progressive Loading**: Show low-quality placeholder while loading

## Security Considerations

### Implemented
- **File Type Validation**: Only allow image MIME types
- **File Size Limits**: Prevent DoS via large files
- **Path Sanitization**: Secure file storage paths
- **User Authentication**: Only authenticated users can upload

### Additional Recommendations
- **Virus Scanning**: Scan uploaded files for malware
- **Rate Limiting**: Limit upload frequency per user
- **Content Moderation**: Automatic image content filtering
- **EXIF Stripping**: Remove metadata from uploaded images

## Summary

The platform now supports comprehensive image and GIF functionality in comments:
- ✅ Image uploads in both regular and group post comments
- ✅ GIF support with preserved animation
- ✅ Real-time preview and validation
- ✅ Responsive design and user-friendly interface
- ✅ Secure file handling and validation
- ✅ Backward compatibility with text-only comments

All changes maintain existing functionality while adding powerful new media capabilities to enhance user engagement.
