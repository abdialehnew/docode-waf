import React, { createContext, useState, useContext, useEffect } from 'react';
import api from '../services/api';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if token exists and validate it
    if (token) {
      validateToken();
    } else {
      setLoading(false);
    }
  }, []);

  const validateToken = async () => {
    try {
      const response = await api.get('/auth/profile');
      setUser(response.data);
    } catch (error) {
      // Token invalid or expired
      logout();
    } finally {
      setLoading(false);
    }
  };

  const login = async (username, password, turnstileToken = '') => {
    const response = await api.post('/auth/login', { 
      username, 
      password,
      turnstile_token: turnstileToken 
    });
    const { token: newToken, admin } = response.data;
    
    localStorage.setItem('token', newToken);
    setToken(newToken);
    setUser(admin);
    
    return response.data;
  };

  const register = async (userData) => {
    const response = await api.post('/auth/register', userData);
    return response.data;
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setUser(null);
  };

  const requestPasswordReset = async (email) => {
    const response = await api.post('/auth/request-reset', { email });
    return response.data;
  };

  const resetPassword = async (token, newPassword) => {
    const response = await api.post('/auth/reset-password', {
      token,
      new_password: newPassword
    });
    return response.data;
  };

  const value = {
    user,
    token,
    loading,
    login,
    register,
    logout,
    requestPasswordReset,
    resetPassword,
    isAuthenticated: !!token && !!user
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
