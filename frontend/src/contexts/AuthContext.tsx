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
  const verifyAbortRef = useRef<AbortController | null>(null);
  const sessionGenRef = useRef(0);

  const cancelPendingVerification = () => {
    verifyAbortRef.current?.abort();
    verifyAbortRef.current = null;
    sessionGenRef.current += 1;
  };

  // Restore session from localStorage on mount (re-runs under React StrictMode with abort).
  useEffect(() => {
    const controller = new AbortController();
    verifyAbortRef.current = controller;
    const generation = ++sessionGenRef.current;
    const { signal } = controller;

    let isActive = true;
    const storedToken = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');

    const loadingFallbackTimer = window.setTimeout(() => {
      if (!isActive || generation !== sessionGenRef.current) return;
      setIsLoading(false);
      if (localStorage.getItem('token')) {
        setAuthStatus('authenticated');
      } else {
        setAuthStatus('guest');
      }
    }, 12000);

    if (!storedToken) {
      setAuthStatus('guest');
      setIsLoading(false);
      return () => {
        isActive = false;
        controller.abort();
        window.clearTimeout(loadingFallbackTimer);
      };
    }

    try {
      const parsedUser = storedUser ? JSON.parse(storedUser) : null;
      setToken(storedToken);
      setUser(parsedUser);

      const tokenAtRequestStart = storedToken;

      authApi
        .getCurrentUser({ signal })
        .then((response) => {
          if (!isActive || generation !== sessionGenRef.current) return;
          if (localStorage.getItem('token') !== tokenAtRequestStart) return;
          setUser(response.data.data);
          localStorage.setItem('user', JSON.stringify(response.data.data));
          setAuthStatus('authenticated');
        })
        .catch(() => {
          if (!isActive || generation !== sessionGenRef.current) return;
          if (signal.aborted) return;
          if (localStorage.getItem('token') !== tokenAtRequestStart) return;
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          setToken(null);
          setUser(null);
          setAuthStatus('guest');
        })
        .finally(() => {
          if (!isActive || generation !== sessionGenRef.current) return;
          setIsLoading(false);
        });
    } catch {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      setToken(null);
      setUser(null);
      setAuthStatus('guest');
      setIsLoading(false);
    }

    return () => {
      isActive = false;
      controller.abort();
      window.clearTimeout(loadingFallbackTimer);
    };
  }, []);

  const login = async (username: string, password: string) => {
    cancelPendingVerification();

    const response = await authApi.login(username, password);
    const { token: newToken, user: newUser } = response.data;

    setToken(newToken);
    setUser(newUser);
    setAuthStatus('authenticated');
    setIsLoading(false);
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
  };

  const logout = () => {
    cancelPendingVerification();
    setToken(null);
    setUser(null);
    setAuthStatus('guest');
    setIsLoading(false);
    localStorage.removeItem('token');
    localStorage.removeItem('user');

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
