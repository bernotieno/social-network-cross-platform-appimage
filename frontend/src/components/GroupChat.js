import React, { useState, useEffect, useRef } from 'react';
import { groupAPI } from '@/utils/api';
import { useAuth } from '@/hooks/useAuth';
import { initializeSocket, getSocket, subscribeToMessages, subscribeToTypingStatus, joinChatRoom, leaveChatRoom, sendTypingStatus } from '@/utils/socket';
import { getUserProfilePictureUrl, getFallbackAvatar } from '@/utils/images';
import { useAlert } from '@/contexts/AlertContext';
import styles from '@/styles/GroupChat.module.css';
import emojis from "@/components/emojis";
import stylesB from '@/styles/Chat.module.css'

export default function GroupChat({ groupId, isVisible }) {
  const { showError } = useAlert();
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [typingUsers, setTypingUsers] = useState(new Map()); // Map of userId -> userInfo for users who are typing
  const [isTyping, setIsTyping] = useState(false);
  const { user } = useAuth();
  const messagesEndRef = useRef(null);
  const chatContainerRef = useRef(null);
  const typingTimeoutRef = useRef(null);

  // Scroll to bottom of messages
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  // Fetch group messages
  const fetchMessages = async () => {
    if (!groupId) return;

    setIsLoading(true);
    try {
      const response = await groupAPI.getGroupMessages(groupId);
      setMessages(response.data.data.messages || []);
    } catch (error) {
      console.error('Error fetching group messages:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Send message
  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (!newMessage.trim() || isSending) return;

    const messageContent = newMessage.trim();
    setIsSending(true);

    // Stop typing indicator when sending message
    if (isTyping) {
      setIsTyping(false);
      const roomId = `group-${groupId}`;
      sendTypingStatus(roomId, false);
    }

    // Clear typing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Create optimistic message object
    const optimisticMessage = {
      id: `temp-${Date.now()}`, // Temporary ID
      content: messageContent,
      senderId: user?.id,
      groupId: groupId,
      createdAt: new Date().toISOString(),
      sender: {
        id: user?.id,
        fullName: user?.fullName,
        username: user?.username,
        profilePicture: user?.profilePicture
      }
    };

    // Add message to local state immediately for instant feedback
    setMessages(prev => [optimisticMessage, ...prev]);
    setNewMessage('');

    try {
      // Send to backend
      const response = await groupAPI.sendGroupMessage(groupId, messageContent);

      // Replace optimistic message with real message from server
      if (response.data && response.data.data && response.data.data.message) {
        const realMessage = response.data.data.message;
        setMessages(prev => prev.map(msg =>
          msg.id === optimisticMessage.id ? realMessage : msg
        ));
      } else {
        // If no real message returned, just remove the optimistic flag
        setMessages(prev => prev.map(msg =>
          msg.id === optimisticMessage.id ? { ...msg, id: `sent-${Date.now()}` } : msg
        ));
      }
    } catch (error) {
      console.error('Error sending message:', error);
      // Remove optimistic message on error
      setMessages(prev => prev.filter(msg => msg.id !== optimisticMessage.id));
      // Restore message content
      setNewMessage(messageContent);
      showError('Failed to send message. Please try again.');
    } finally {
      setIsSending(false);
    }
  };

  // Handle emoji insertion
  const insertEmoji = (emoji) => {
    handleTyping(newMessage + emoji);
  };

  // Handle typing status
  const handleTyping = (value) => {
    setNewMessage(value);

    if (!groupId) return;

    const roomId = `group-${groupId}`;

    if (value.trim() && !isTyping) {
      // User started typing
      setIsTyping(true);
      sendTypingStatus(roomId, true);
    } else if (!value.trim() && isTyping) {
      // User stopped typing (cleared input)
      setIsTyping(false);
      sendTypingStatus(roomId, false);
    }

    // Clear existing timeout
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    // Set timeout to stop typing indicator after 3 seconds of inactivity
    if (value.trim()) {
      typingTimeoutRef.current = setTimeout(() => {
        setIsTyping(false);
        sendTypingStatus(roomId, false);
      }, 3000);
    }
  };

  // Initialize WebSocket connection
  useEffect(() => {
    const socket = initializeSocket();
    if (!socket) {
      console.error('Failed to initialize WebSocket for group chat');
    }
  }, []);

  // WebSocket message handler
  useEffect(() => {
    if (!groupId) return;

    const roomId = `group-${groupId}`;

    // Join the chat room
    console.log('Joining group chat room:', roomId);
    joinChatRoom(roomId);

    // Subscribe to messages
    const unsubscribeMessages = subscribeToMessages((data) => {
      console.log('GroupChat received WebSocket message:', data);

      // Extract payload from WebSocket message (same as normal chat)
      const payload = data.payload || data;
      const messageRoomId = payload.roomId || data.roomId;
      const messageData = payload.message || data.message;

      console.log('Extracted payload:', payload);
      console.log('Message room ID:', messageRoomId);
      console.log('Current room ID:', roomId);

      if (messageRoomId === roomId) {
        console.log('Processing group message data:', messageData);

        const newMsg = {
          id: messageData.id || `ws-${Date.now()}-${messageData.sender}`,
          content: messageData.content,
          senderId: messageData.sender,
          groupId: messageData.groupId || groupId,
          createdAt: messageData.timestamp,
          sender: messageData.senderInfo || {
            id: messageData.sender,
            fullName: 'Unknown User', // Fallback for WebSocket messages
            username: 'unknown'
          }
        };

        console.log('Created new message object:', newMsg);

        // Handle message deduplication (same logic as normal chat)
        setMessages(prev => {
          console.log("Processing WebSocket message:", newMsg);
          console.log("Current messages:", prev);

          // Check if message already exists to prevent duplicates
          // First check for exact optimistic message match
          const optimisticMessageIndex = prev.findIndex(msg => {
            const isMatch = msg.id && msg.id.startsWith('temp-') &&
              msg.content === newMsg.content &&
              msg.senderId === newMsg.senderId;
            console.log("Checking optimistic message:", msg, "Match:", isMatch);
            return isMatch;
          });

          if (optimisticMessageIndex !== -1) {
            // Replace optimistic message with confirmed message
            console.log("Replacing optimistic message with confirmed message at index:", optimisticMessageIndex);
            const updatedMessages = [...prev];
            updatedMessages[optimisticMessageIndex] = newMsg;
            console.log("Updated messages:", updatedMessages);
            return updatedMessages;
          }

          // Check if message already exists (for regular duplicates)
          const messageExists = prev.some(msg =>
            !msg.id?.startsWith('temp-') &&
            msg.content === newMsg.content &&
            msg.senderId === newMsg.senderId &&
            Math.abs(new Date(msg.createdAt) - new Date(newMsg.createdAt)) < 5000 // Within 5 seconds
          );

          if (!messageExists) {
            console.log("Adding new message:", newMsg);
            return [newMsg, ...prev];
          } else {
            console.log("Message already exists, skipping");
            return prev;
          }
        });
      }
    });

    // Subscribe to typing status updates
    const unsubscribeTyping = subscribeToTypingStatus((typingData) => {
      console.log('Received typing status:', typingData);
      if (typingData.roomId === roomId && typingData.userId !== user?.id) {
        setTypingUsers(prev => {
          const newMap = new Map(prev);
          if (typingData.isTyping) {
            // Add user with their info
            newMap.set(typingData.userId, typingData.userInfo || {
              id: typingData.userId,
              fullName: 'Unknown User',
              username: 'unknown'
            });
          } else {
            // Remove user
            newMap.delete(typingData.userId);
          }
          return newMap;
        });
      }
    });

    return () => {
      // Leave room and unsubscribe
      leaveChatRoom(roomId);
      unsubscribeMessages();
      unsubscribeTyping();
    };
  }, [groupId, user?.id]);

  // Fetch messages when component mounts or groupId changes
  useEffect(() => {
    if (isVisible && groupId) {
      fetchMessages();
    }
  }, [groupId, isVisible]);

  // Scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Cleanup typing timeout on unmount
  useEffect(() => {
    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, []);


  const emojiList = emojis();
  const [showEmojis, setShowEmojis] = useState(false);


  if (!isVisible) return null;


  return (
    <div className={styles.groupChat}>
      <div className={styles.chatHeader}>
        <h3>Group Chat</h3>
      </div>

      <div className={styles.messagesContainer} ref={chatContainerRef}>
        {isLoading ? (
          <div className={styles.loading}>Loading messages...</div>
        ) : messages.length === 0 ? (
          <div className={styles.noMessages}>
            No messages yet. Start the conversation!
          </div>
        ) : (
          <div className={styles.messagesList}>
            {messages.slice().reverse().map((message, index) => (
              <div
                key={message.id || `temp-${index}-${message.createdAt}`}
                className={`${styles.message} ${
                  message.senderId === user?.id ? styles.ownMessage : styles.otherMessage
                }`}
              >
                <div className={styles.messageHeader}>
                  <img
                    src={message.sender?.profilePicture ? getUserProfilePictureUrl(message.sender) : getFallbackAvatar(message.sender)}
                    alt={message.sender?.fullName}
                    className={styles.senderAvatar}
                    onError={(e) => {
                      e.target.src = getFallbackAvatar(message.sender);
                    }}
                  />
                  <span className={styles.senderName}>
                    {message.senderId === user?.id ? 'You' : message.sender?.fullName}
                  </span>
                  <span className={styles.messageTime}>
                    {new Date(message.createdAt).toLocaleTimeString([], {
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </span>
                </div>
                <div className={styles.messageContent}>
                  {message.content}
                </div>
              </div>
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}

        {/* Typing indicator */}
        {typingUsers.size > 0 && (
          <div className={styles.typingIndicator}>
            {(() => {
              const typingUsersList = Array.from(typingUsers.values());
              if (typingUsersList.length === 1) {
                return `${typingUsersList[0].fullName || typingUsersList[0].username || 'Someone'} is typing...`;
              } else if (typingUsersList.length === 2) {
                const names = typingUsersList.map(u => u.fullName || u.username || 'Someone');
                return `${names[0]} and ${names[1]} are typing...`;
              } else {
                const firstTwo = typingUsersList.slice(0, 2).map(u => u.fullName || u.username || 'Someone');
                const remaining = typingUsersList.length - 2;
                return `${firstTwo.join(', ')} and ${remaining} other${remaining > 1 ? 's' : ''} are typing...`;
              }
            })()}
          </div>
        )}
      </div>

      <form onSubmit={handleSendMessage} className={styles.messageForm}>
        {showEmojis && (
            <div className={stylesB.emojiBarGroup}>
              {emojiList.map((emoji,key) => (
                  <button
                      key={key}
                      type="button"
                      className={stylesB.emojiButton}
                      onClick={() => insertEmoji(emoji)}
                  >
                    {emoji}
                  </button>
              ))}
            </div>
        )}

        <div className={styles.inputContainer}>
          <input
            type="text"
            value={newMessage}
            onChange={(e) => handleTyping(e.target.value)}
            placeholder="Type a message..."
            className={styles.messageInput}
            disabled={isSending}
          />

          {/*toggle emoji box*/}
          <button
              type="button"
              onClick={() => setShowEmojis(!showEmojis)}
              className={stylesB.emojiToggleButton}
          >
            {showEmojis ? 'hide emojis' : 'show more emojis'}
          </button>

          <button
            type="submit"
            disabled={!newMessage.trim() || isSending}
            className={styles.sendButton}
          >
            {isSending ? 'Sending...' : 'Send'}
          </button>
        </div>
      </form>
    </div>
  );
}
