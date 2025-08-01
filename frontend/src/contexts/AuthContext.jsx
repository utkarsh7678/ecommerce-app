import { createContext, useContext, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  // Check if user is logged in on initial load
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      // Set the default Authorization header for all axios requests
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      // Fetch user data
      const fetchUser = async () => {
        try {
          const response = await axios.get('/api/users/me');
          setUser(response.data);
        } catch (error) {
          console.error('Failed to fetch user data', error);
          logout();
        } finally {
          setLoading(false);
        }
      };
      
      fetchUser();
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (username, password) => {
    try {
      const response = await axios.post('/api/users/login', { 
        username: username.trim(), 
        password: password 
      });
      
      const { token, user_id } = response.data;
      
      if (!token) {
        throw new Error('No token received');
      }
      
      // Store token in localStorage
      localStorage.setItem('token', token);
      
      // Set the default Authorization header
      axios.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      // Fetch user data
      const userResponse = await axios.get(`/api/users/me`);
      
      setUser({
        id: user_id,
        username: username,
        ...userResponse.data
      });
      
      return { success: true };
    } catch (error) {
      console.error('Login failed:', error);
      const errorMessage = error.response?.data?.error || 
                         error.response?.data?.message || 
                         'Login failed. Please check your credentials and try again.';
      
      return { 
        success: false, 
        error: errorMessage
      };
    }
  };

  const register = async (username, password) => {
    try {
      console.log('Sending registration request with:', { username, password });
      const response = await axios.post('/api/users', { 
        username: username.trim(),
        password: password
      }, {
        headers: {
          'Content-Type': 'application/json'
        }
      });
      console.log('Registration response:', response.data);
      return { success: true };
    } catch (error) {
      console.error('Registration failed:', {
        message: error.message,
        response: error.response?.data,
        status: error.response?.status,
        headers: error.response?.headers
      });
      
      let errorMessage = 'Registration failed. Please try again.';
      
      if (error.response) {
        // The request was made and the server responded with a status code
        // that falls out of the range of 2xx
        if (error.response.data && error.response.data.error) {
          errorMessage = error.response.data.error;
        } else if (error.response.data && error.response.data.details) {
          errorMessage = error.response.data.details;
        } else if (error.response.status === 400) {
          errorMessage = 'Invalid request. Please check your input and try again.';
        } else if (error.response.status === 409) {
          errorMessage = 'Username already exists. Please choose a different username.';
        }
      } else if (error.request) {
        // The request was made but no response was received
        errorMessage = 'No response from server. Please check your connection.';
      }
      
      return { 
        success: false, 
        error: errorMessage
      };
    }
  };

  const logout = () => {
    // Remove token from localStorage
    localStorage.removeItem('token');
    
    // Remove Authorization header
    delete axios.defaults.headers.common['Authorization'];
    
    // Clear user state
    setUser(null);
    
    // Redirect to login
    navigate('/login');
  };

  const value = {
    user,
    isAuthenticated: !!user,
    loading,
    login,
    register,
    logout,
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

export default AuthContext;