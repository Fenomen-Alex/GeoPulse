import { createSignal, onMount } from 'solid-js';

interface AuthStatus {
  authenticated: boolean;
  user?: any;
  error?: string;
}

export const useAuth = () => {
  const [authStatus, setAuthStatus] = createSignal<AuthStatus>({
    authenticated: false,
  });

  const checkAuthStatus = async () => {
    try {
      const res = await fetch('/api/v1/auth/status');
      if (res.ok) {
        const data = await res.json();
        setAuthStatus({
          authenticated: true,
          user: data.user,
        });
      } else {
        setAuthStatus({
          authenticated: false,
        });
      }
    } catch (err) {
      setAuthStatus({
        authenticated: false,
        error: err instanceof Error ? err.message : 'Authentication check failed',
      });
    }
  };

  const login = async () => {
    try {
      // Open Google Auth modal
      const res = await fetch('/api/v1/auth/login');
      if (res.ok) {
        const data = await res.json();
        if (data.authenticated) {
          setAuthStatus({
            authenticated: true,
            user: data.user,
          });
          return data;
        }
      }
      return { success: false, message: 'Login failed' };
    } catch (err) {
      return { success: false, message: err instanceof Error ? err.message : 'Login failed' };
    }
  };

  const logout = async () => {
    try {
      const res = await fetch('/api/v1/auth/logout', { method: 'POST' });
      if (res.ok) {
        setAuthStatus({
          authenticated: false,
        });
        return { success: true };
      }
      return { success: false, message: 'Logout failed' };
    } catch (err) {
      return { success: false, message: err instanceof Error ? err.message : 'Logout failed' };
    }
  };

  onMount(() => {
    checkAuthStatus();
  });

  return {
    authStatus,
    checkAuthStatus,
    login,
    logout,
  };
};