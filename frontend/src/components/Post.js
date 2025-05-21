'use client';

import { useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { formatDistanceToNow } from 'date-fns';
import { useAuth } from '@/hooks/useAuth';
import { postAPI } from '@/utils/api';
import Button from '@/components/Button';
import styles from '@/styles/Post.module.css';

const Post = ({ post, onDelete }) => {
  const { user } = useAuth();
  const [isLiked, setIsLiked] = useState(post.isLikedByCurrentUser || false);
  const [likesCount, setLikesCount] = useState(post.likesCount || 0);
  const [showComments, setShowComments] = useState(false);
  const [comments, setComments] = useState(post.comments || []);
  const [newComment, setNewComment] = useState('');
  const [isSubmittingComment, setIsSubmittingComment] = useState(false);
  
  const isOwnPost = user?.id === post.author.id;
  
  const handleLikeToggle = async () => {
    try {
      if (isLiked) {
        await postAPI.unlikePost(post.id);
        setIsLiked(false);
        setLikesCount(prev => prev - 1);
      } else {
        await postAPI.likePost(post.id);
        setIsLiked(true);
        setLikesCount(prev => prev + 1);
      }
    } catch (error) {
      console.error('Error toggling like:', error);
    }
  };
  
  const handleCommentToggle = async () => {
    if (!showComments) {
      try {
        const response = await postAPI.getComments(post.id);
        setComments(response.data.comments || []);
      } catch (error) {
        console.error('Error fetching comments:', error);
      }
    }
    setShowComments(!showComments);
  };
  
  const handleCommentSubmit = async (e) => {
    e.preventDefault();
    
    if (!newComment.trim()) return;
    
    try {
      setIsSubmittingComment(true);
      const response = await postAPI.addComment(post.id, newComment);
      setComments(prev => [...prev, response.data]);
      setNewComment('');
    } catch (error) {
      console.error('Error adding comment:', error);
    } finally {
      setIsSubmittingComment(false);
    }
  };
  
  const handleDeleteComment = async (commentId) => {
    try {
      await postAPI.deleteComment(post.id, commentId);
      setComments(prev => prev.filter(comment => comment.id !== commentId));
    } catch (error) {
      console.error('Error deleting comment:', error);
    }
  };
  
  const handleDeletePost = async () => {
    if (window.confirm('Are you sure you want to delete this post?')) {
      try {
        await postAPI.deletePost(post.id);
        if (onDelete) {
          onDelete(post.id);
        }
      } catch (error) {
        console.error('Error deleting post:', error);
      }
    }
  };
  
  return (
    <div className={styles.post}>
      <div className={styles.postHeader}>
        <Link href={`/profile/${post.author.id}`} className={styles.authorInfo}>
          {post.author.profilePicture ? (
            <Image 
              src={post.author.profilePicture} 
              alt={post.author.username} 
              width={40} 
              height={40} 
              className={styles.authorAvatar}
            />
          ) : (
            <div className={styles.authorAvatarPlaceholder}>
              {post.author.username?.charAt(0).toUpperCase() || 'U'}
            </div>
          )}
          <div>
            <h3 className={styles.authorName}>{post.author.fullName}</h3>
            <p className={styles.postTime}>
              {formatDistanceToNow(new Date(post.createdAt), { addSuffix: true })}
            </p>
          </div>
        </Link>
        
        {isOwnPost && (
          <div className={styles.postActions}>
            <button 
              className={styles.actionButton} 
              onClick={handleDeletePost}
              aria-label="Delete post"
            >
              🗑️
            </button>
          </div>
        )}
      </div>
      
      <div className={styles.postContent}>
        <p className={styles.postText}>{post.content}</p>
        
        {post.image && (
          <div className={styles.postImage}>
            <Image 
              src={post.image} 
              alt="Post image" 
              fill 
              style={{ objectFit: 'cover' }}
            />
          </div>
        )}
      </div>
      
      <div className={styles.postFooter}>
        <div className={styles.postStats}>
          {likesCount > 0 && (
            <span className={styles.likesCount}>
              {likesCount} {likesCount === 1 ? 'like' : 'likes'}
            </span>
          )}
          
          {post.commentsCount > 0 && (
            <span className={styles.commentsCount}>
              {post.commentsCount} {post.commentsCount === 1 ? 'comment' : 'comments'}
            </span>
          )}
        </div>
        
        <div className={styles.postButtons}>
          <button 
            className={`${styles.postButton} ${isLiked ? styles.liked : ''}`}
            onClick={handleLikeToggle}
          >
            {isLiked ? '❤️' : '🤍'} Like
          </button>
          
          <button 
            className={styles.postButton}
            onClick={handleCommentToggle}
          >
            💬 Comment
          </button>
        </div>
        
        {showComments && (
          <div className={styles.commentsSection}>
            <form onSubmit={handleCommentSubmit} className={styles.commentForm}>
              <input
                type="text"
                placeholder="Write a comment..."
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                className={styles.commentInput}
                disabled={isSubmittingComment}
              />
              <Button 
                type="submit" 
                variant="primary" 
                size="small"
                disabled={isSubmittingComment || !newComment.trim()}
              >
                Post
              </Button>
            </form>
            
            <div className={styles.commentsList}>
              {comments.length === 0 ? (
                <p className={styles.noComments}>No comments yet. Be the first to comment!</p>
              ) : (
                comments.map(comment => (
                  <div key={comment.id} className={styles.comment}>
                    <Link href={`/profile/${comment.author.id}`} className={styles.commentAuthor}>
                      {comment.author.profilePicture ? (
                        <Image 
                          src={comment.author.profilePicture} 
                          alt={comment.author.username} 
                          width={32} 
                          height={32} 
                          className={styles.commentAvatar}
                        />
                      ) : (
                        <div className={styles.commentAvatarPlaceholder}>
                          {comment.author.username?.charAt(0).toUpperCase() || 'U'}
                        </div>
                      )}
                    </Link>
                    
                    <div className={styles.commentContent}>
                      <div className={styles.commentBubble}>
                        <h4 className={styles.commentAuthorName}>{comment.author.fullName}</h4>
                        <p className={styles.commentText}>{comment.content}</p>
                      </div>
                      
                      <div className={styles.commentMeta}>
                        <span className={styles.commentTime}>
                          {formatDistanceToNow(new Date(comment.createdAt), { addSuffix: true })}
                        </span>
                        
                        {(user?.id === comment.author.id || isOwnPost) && (
                          <button 
                            className={styles.deleteCommentButton}
                            onClick={() => handleDeleteComment(comment.id)}
                            aria-label="Delete comment"
                          >
                            Delete
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Post;
