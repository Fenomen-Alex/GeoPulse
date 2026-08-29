import { createSignal, onMount } from 'solid-js';

interface AuthStatus {
  authenticated: boolean;
  loading: boolean;
  user?: any;
  error?: string;
}

export const useAuth = () => {
  const [authStatus, setAuthStatus] = createSignal<AuthStatus>({
    authenticated: false,
    loading: true,
  });

  const isTestMode = () =>
    import.meta.env.DEV &&
    (import.meta.env.VITE_TEST_MODE === 'true' || import.meta.env.VITE_NODE_ENV === 'test');

  const checkAuthStatus = async () => {
    if (isTestMode()) {
      setAuthStatus({
        authenticated: true,
        loading: false,
        user: { id: 'test-user', name: 'Test User', email: 'test@geopulse.local' },
      });
      return;
    }

    try {
      const res = await fetch('/api/v1/auth/status');
      if (res.ok) {
        const data = await res.json();
        setAuthStatus({
          authenticated: data.authenticated === true,
          loading: false,
          user: data.user,
        });
      } else {
        setAuthStatus({
          authenticated: false,
          loading: false,
        });
      }
    } catch (err) {
      setAuthStatus({
        authenticated: false,
        loading: false,
        error: err instanceof Error ? err.message : 'Authentication check failed',
      });
    }
  };

  const login = async () => {
    if (isTestMode()) {
      setAuthStatus({
        authenticated: true,
        loading: false,
        user: { id: 'test-user', name: 'Test User', email: 'test@geopulse.local' },
      });
      return { success: true };
    }

    window.location.href = '/api/v1/auth/login';
    return { success: true };
  };

  const logout = async () => {
    if (isTestMode()) return { success: true };

    try {
      const res = await fetch('/api/v1/auth/logout', { method: 'POST' });
      if (res.ok) {
        setAuthStatus({
          authenticated: false,
          loading: false,
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
