import type { Component } from 'solid-js';
import { useAuth } from './services/api';
import { AppShell } from './components/AppShell';
import { LandingView } from './views/LandingView';
import { AppView } from './views/AppView';

const App: Component = () => {
  const { authStatus, logout } = useAuth();

  const handleLogout = async () => {
    const result = await logout();
    if (result.success) window.location.href = '/';
  };

  return (
    <AppShell>
      {authStatus().loading ? (
        <div class="flex min-h-screen items-center justify-center bg-zinc-950 text-zinc-400">
          <div class="flex items-center gap-3 text-sm">
            <span class="h-2.5 w-2.5 animate-pulse rounded-full bg-cyan-400" />
            Loading GeoPulse…
          </div>
        </div>
      ) : authStatus().authenticated ? (
        <AppView user={authStatus().user} onLogout={handleLogout} />
      ) : (
        <LandingView />
      )}
    </AppShell>
  );
};

export default App;
