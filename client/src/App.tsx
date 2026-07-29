import type { Component } from 'solid-js';
import { useAuth } from './services/api';
import { AppShell } from './components/AppShell';
import { LandingView } from './views/LandingView';
import { AppView } from './views/AppView';

const App: Component = () => {
  const { authStatus } = useAuth();

  return (
    <AppShell>
      {authStatus().authenticated ? <AppView user={authStatus().user} /> : <LandingView />}
    </AppShell>
  );
};

export default App;
