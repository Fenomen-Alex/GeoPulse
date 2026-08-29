import { createEffect, onCleanup, type Component, Show } from 'solid-js';

interface AuthChoiceModalProps {
  open: boolean;
  onClose: () => void;
}

const startGoogleAuth = () => {
  window.location.href = '/api/v1/auth/login';
};

export const AuthChoiceModal: Component<AuthChoiceModalProps> = (props) => {
  let dialog!: HTMLElement;

  createEffect(() => {
    if (!props.open) return;

    const previousFocus = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') props.onClose();
    };

    document.body.style.overflow = 'hidden';
    document.addEventListener('keydown', closeOnEscape);
    queueMicrotask(() => dialog?.querySelector<HTMLElement>('button')?.focus());

    onCleanup(() => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener('keydown', closeOnEscape);
      previousFocus?.focus();
    });
  });

  return (
    <Show when={props.open}>
      <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-6 backdrop-blur-md"
        role="presentation"
        onClick={(event) => {
          if (event.target === event.currentTarget) props.onClose();
        }}
      >
        <section
          ref={dialog}
          class="relative w-full max-w-md rounded-2xl border border-zinc-700/80 bg-zinc-900 p-7 shadow-2xl shadow-cyan-950/40"
          role="dialog"
          aria-modal="true"
          aria-labelledby="auth-choice-title"
        >
          <button
            type="button"
            aria-label="Close sign in dialog"
            onClick={props.onClose}
            class="absolute right-4 top-4 text-xl leading-none text-zinc-500 transition-colors hover:text-zinc-200"
          >
            ×
          </button>

          <div class="mb-6 flex items-start gap-3">
            <div class="mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-cyan-400/10 text-cyan-400">
              <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 2 2 7l10 5 10-5-10-5Zm-10 10 10 5 10-5M2 17l10 5 10-5" />
              </svg>
            </div>
            <div>
              <h2 id="auth-choice-title" class="text-xl font-semibold text-zinc-100">Enter the GeoPulse workbench</h2>
              <p class="mt-1 text-sm leading-6 text-zinc-400">Choose how you want to continue. Your Google account keeps your workbench and daily quota secure.</p>
            </div>
          </div>

          <div class="space-y-3">
            <button
              type="button"
              onClick={startGoogleAuth}
              class="flex w-full items-center justify-between rounded-xl bg-cyan-500 px-4 py-3 text-left font-medium text-zinc-950 transition-colors hover:bg-cyan-400"
            >
              <span>Create a free account</span>
              <span aria-hidden="true">→</span>
            </button>
            <button
              type="button"
              onClick={startGoogleAuth}
              class="flex w-full items-center justify-between rounded-xl border border-zinc-700 px-4 py-3 text-left font-medium text-zinc-200 transition-colors hover:border-cyan-400/60 hover:text-cyan-300"
            >
              <span>Sign in with Google</span>
              <span aria-hidden="true">→</span>
            </button>
          </div>

          <p class="mt-5 text-center text-xs text-zinc-500">Free access includes 15 spatial runs per day. No credit card required.</p>
        </section>
      </div>
    </Show>
  );
};
