import type { Component } from 'solid-js';

export const AppShell: Component<{ children: any }> = (props) => {
  return <div class="min-h-screen">{props.children}</div>;
};