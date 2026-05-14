// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { createContext, useContext, useState, useEffect, ReactNode, useRef } from 'react';
import { authApi, User } from '../services/api';

type AuthStatus = 'checking' | 'authenticated' | 'guest';

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
  isAuthenticated: boolean;
  authStatus: AuthStatus;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [authStatus, setAuthStatus] = useState<AuthStatus>('checking');
  const initCompletedRef = useRef(false);

  // Load user from localStorage on mount
  useEffect(() => {
    if (initCompletedRef.current) {
      return;
    }
    initCompletedRef.current = true;

    let isActive = true;
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    const loadingFallbackTimer = window.setTimeout(() => {
      if (isActive) {
        setIsLoading(false);
        setAuthStatus(storedToken ? 'authenticated' : 'guest');
      }
    }, 12000);

    if (storedToken) {
      try {
        const parsedUser = storedUser ? JSON.parse(storedUser) : null;
        setToken(storedToken);
        setUser(parsedUser);

        // Token used for this verification; ignore late responses after login replaces it.
        const tokenAtRequestStart = storedToken;

        // Verify token by fetching current user
        authApi.getCurrentUser()
          .then((response) => {
            if (!isActive) return;
            if (localStorage.getItem('token') !== tokenAtRequestStart) return;
            setUser(response.data.data);
            localStorage.setItem('user', JSON.stringify(response.data.data));
            setAuthStatus('authenticated');
          })
          .catch(() => {
            if (!isActive) return;
            if (localStorage.getItem('token') !== tokenAtRequestStart) return;
            // Token invalid, clear storage
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            setToken(null);
            setUser(null);
            setAuthStatus('guest');
          })
          .finally(() => {
            if (!isActive) return;
            setIsLoading(false);
          });
      } catch (error) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setToken(null);
        setUser(null);
        setAuthStatus('guest');
        if (isActive) {
          setIsLoading(false);
        }
      }
    } else {
      setAuthStatus('guest');
      setIsLoading(false);
    }

    return () => {
      isActive = false;
      window.clearTimeout(loadingFallbackTimer);
    };
  }, []);

  const login = async (username: string, password: string) => {
    const response = await authApi.login(username, password);
    const { token: newToken, user: newUser } = response.data;

    setToken(newToken);
    setUser(newUser);
    setAuthStatus('authenticated');
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    setAuthStatus('guest');
    localStorage.removeItem('token');
    localStorage.removeItem('user');

    // Call logout API (optional, for server-side cleanup if needed)
    authApi.logout().catch(() => {
      // Ignore errors on logout
    });
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        login,
        logout,
        isLoading,
        isAuthenticated: !!user && !!token,
        authStatus,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

