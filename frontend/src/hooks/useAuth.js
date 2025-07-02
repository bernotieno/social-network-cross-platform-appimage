'use client';

import { createContext, useContext, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  setAuth,
  getUser,
  getToken,
  logout as logoutAuth,
  isAuthenticated as checkAuth,
  updateUserData as updateStoredUserData
} from '@/utils/auth';
import { authAPI } from '@/utils/api';
import { disconnectSocket, initializeSocket } from '@/utils/socket';

// Create auth context
const AuthContext = createContext({
  user: null,
  isAuthenticated: false,
  isLoading: true,
  login: async () => {},
  register: async () => {},
  logout: () => {},
  updateUserData: () => {},
});

// Auth provider component
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  // Check if user is authenticated on mount
  useEffect(() => {
    const checkAuthentication = () => {
      const userData = getUser();
      if (userData) {
        setUser(userData);
        // Initialize WebSocket connection for authenticated users
        try {
          initializeSocket();
        } catch (error) {
          console.warn('Failed to initialize WebSocket:', error);
        }
      }
      setIsLoading(false);
    };

    checkAuthentication();
  }, []);

  // Login function
  const login = async (email, password) => {
    try {
      setIsLoading(true);
      const response = await authAPI.login(email, password);

      // Check if response.data exists and has the expected structure
      if (!response.data || !response.data.data) {
        throw new Error('Invalid response structure from login API');
      }

      const token = response.data.data.token;
      const user = response.data.data.user;

      // Validate token
      if (!token || token === '') {
        throw new Error(`No token received from login API. Got: "${token}"`);
      }

      if (!user) {
        throw new Error(`No user data received from login API. Got: ${user}`);
      }

      // Clear any existing sessions first
      await authAPI.logout(); // This ensures old sessions are invalidated

      // Set new session
      setAuth(token, user);
      setUser(user);

      // Initialize WebSocket connection after successful login
      try {
        initializeSocket();
      } catch (error) {
        console.warn('Failed to initialize WebSocket - Real-time features will be disabled:', error);
      }

      return { success: true };
    } catch (error) {
      console.error('Login error:', error);
      return {
        success: false,
        error: error.response?.data?.message || 'Login failed'
      };
    } finally {
      setIsLoading(false);
    }
  };

  // Register function
  const register = async (userData) => {
    try {
      setIsLoading(true);
      console.log('Sending registration data:', userData);
      const response = await authAPI.register(userData);
      console.log('Registration response:', response.data);

      // Parse response data - backend returns data in response.data.data
      const token = response.data.data.token;
      const user = response.data.data.user;

      setAuth(token, user);
      setUser(user);

      // Initialize WebSocket connection after successful registration
      try {
        initializeSocket();
      } catch (error) {
        console.warn('Failed to initialize WebSocket after registration:', error);
      }

      return { success: true, user };
    } catch (error) {
      console.error('Registration error:', error);
      return {
        success: false,
        error: error.response?.data?.message || 'Registration failed'
      };
    } finally {
      setIsLoading(false);
    }
  };

  // Logout function
  const logout = async () => {
    try {
      // Call logout API if authenticated
      if (checkAuth()) {
        await authAPI.logout();
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Disconnect socket
      disconnectSocket();

      // Clear auth data
      logoutAuth();
      setUser(null);

      // Redirect to login page
      router.push('/auth/login');
    }
  };

  // Update user data
  const updateUserData = (userData) => {
    setUser(prevUser => ({ ...prevUser, ...userData }));
    updateStoredUserData(userData);
  };

  // Context value
  const value = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    register,
    logout,
    updateUserData,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use auth context
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export default useAuth;
