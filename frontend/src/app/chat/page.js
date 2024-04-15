'use client';

import { useState, useEffect, useRef } from 'react';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { userAPI } from '@/utils/api';
import { 
  getSocket, 
  joinChatRoom, 
  leaveChatRoom, 
  sendMessage, 
  subscribeToMessages 
} from '@/utils/socket';
import Button from '@/components/Button';
import ProtectedRoute from '@/components/ProtectedRoute';
import styles from '@/styles/Chat.module.css';

export default function Chat() {
  const { user } = useAuth();
  const [contacts, setContacts] = useState([]);
  const [selectedContact, setSelectedContact] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  
  const messagesEndRef = useRef(null);
  
  // Fetch contacts (followers/following)
  useEffect(() => {
    const fetchContacts = async () => {
      try {
        setIsLoading(true);
        
        // Get user's following list as contacts
        const response = await userAPI.getFollowing(user?.id);
        setContacts(response.data.following || []);
      } catch (error) {
        console.error('Error fetching contacts:', error);
      } finally {
        setIsLoading(false);
      }
    };
    
    if (user) {
      fetchContacts();
    }
  }, [user]);
  
  // Handle chat room subscription
  useEffect(() => {
    if (!selectedContact) return;
    
    // Create a room ID (combination of user IDs sorted alphabetically)
    const roomId = [user.id, selectedContact.id].sort().join('-');
    
    // Join the chat room
    joinChatRoom(roomId);
    
    // Subscribe to messages
    const unsubscribe = subscribeToMessages((data) => {
      if (data.roomId === roomId) {
        setMessages(prev => [...prev, data.message]);
      }
    });
    
    // Fetch previous messages
    const fetchMessages = async () => {
      try {
        // This would be an API call to get previous messages
        // For now, we'll just set an empty array
        setMessages([]);
      } catch (error) {
        console.error('Error fetching messages:', error);
      }
    };
    
    fetchMessages();
    
    // Clean up on unmount or when changing contacts
    return () => {
      leaveChatRoom(roomId);
      unsubscribe();
    };
  }, [selectedContact, user]);
  
  // Scroll to bottom when messages change
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);
  
  const handleContactSelect = (contact) => {
    setSelectedContact(contact);
  };
  
  const handleSendMessage = (e) => {
    e.preventDefault();
    
    if (!newMessage.trim() || !selectedContact) return;
    
    // Create a room ID (combination of user IDs sorted alphabetically)
    const roomId = [user.id, selectedContact.id].sort().join('-');
    
    // Send message through socket
    sendMessage(roomId, {
      content: newMessage,
      sender: user.id,
      timestamp: new Date().toISOString(),
    });
    
    // Clear input
    setNewMessage('');
  };
  
  return (
    <ProtectedRoute>
      <div className={styles.chatContainer}>
        <div className={styles.chatSidebar}>
          <h2 className={styles.sidebarTitle}>Conversations</h2>
          
          {isLoading ? (
            <div className={styles.loading}>Loading contacts...</div>
          ) : contacts.length === 0 ? (
            <div className={styles.emptyContacts}>
              <p>No contacts found</p>
              <p>Follow people to start chatting with them</p>
            </div>
          ) : (
            <div className={styles.contactsList}>
              {contacts.map(contact => (
                <div 
                  key={contact.id}
                  className={`${styles.contactItem} ${selectedContact?.id === contact.id ? styles.activeContact : ''}`}
                  onClick={() => handleContactSelect(contact)}
                >
                  {contact.profilePicture ? (
                    <Image 
                      src={contact.profilePicture} 
                      alt={contact.username} 
                      width={40} 
                      height={40} 
                      className={styles.contactAvatar}
                    />
                  ) : (
                    <div className={styles.contactAvatarPlaceholder}>
                      {contact.username?.charAt(0).toUpperCase() || 'U'}
                    </div>
                  )}
                  
                  <div className={styles.contactInfo}>
                    <h3 className={styles.contactName}>{contact.fullName}</h3>
                    <p className={styles.contactUsername}>@{contact.username}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
        
        <div className={styles.chatMain}>
          {!selectedContact ? (
            <div className={styles.noChatSelected}>
              <p>Select a contact to start chatting</p>
            </div>
          ) : (
            <>
              <div className={styles.chatHeader}>
                <div className={styles.chatHeaderInfo}>
                  {selectedContact.profilePicture ? (
                    <Image 
                      src={selectedContact.profilePicture} 
                      alt={selectedContact.username} 
                      width={40} 
                      height={40} 
                      className={styles.headerAvatar}
                    />
                  ) : (
                    <div className={styles.headerAvatarPlaceholder}>
                      {selectedContact.username?.charAt(0).toUpperCase() || 'U'}
                    </div>
                  )}
                  
                  <div>
                    <h2 className={styles.headerName}>{selectedContact.fullName}</h2>
                    <p className={styles.headerUsername}>@{selectedContact.username}</p>
                  </div>
                </div>
              </div>
              
              <div className={styles.messagesContainer}>
                {messages.length === 0 ? (
                  <div className={styles.emptyMessages}>
                    <p>No messages yet</p>
                    <p>Send a message to start the conversation</p>
                  </div>
                ) : (
                  <div className={styles.messagesList}>
                    {messages.map((message, index) => (
                      <div 
                        key={index}
                        className={`${styles.messageItem} ${message.sender === user.id ? styles.ownMessage : styles.otherMessage}`}
                      >
                        <div className={styles.messageContent}>
                          {message.content}
                        </div>
                        <div className={styles.messageTime}>
                          {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </div>
                      </div>
                    ))}
                    <div ref={messagesEndRef} />
                  </div>
                )}
              </div>
              
              <form onSubmit={handleSendMessage} className={styles.messageForm}>
                <input
                  type="text"
                  placeholder="Type a message..."
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                  className={styles.messageInput}
                />
                <Button 
                  type="submit" 
                  variant="primary"
                  disabled={!newMessage.trim()}
                >
                  Send
                </Button>
              </form>
            </>
          )}
        </div>
      </div>
    </ProtectedRoute>
  );
}
