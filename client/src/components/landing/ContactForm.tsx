import { type Component, createSignal, Show } from 'solid-js';

export const ContactForm: Component = () => {
  const [submitted, setSubmitted] = createSignal(false);
  const [loading, setLoading] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setLoading(true);

    const form = e.target as HTMLFormElement;
    const data = new FormData(form);

    try {
      const res = await fetch('/api/v1/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: data.get('name'),
          email: data.get('email'),
          subject: data.get('subject'),
          message: data.get('message'),
        }),
      });

      if (res.ok) {
        setSubmitted(true);
        form.reset();
      }
    } catch (err) {
      console.error('Contact form error:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <section class="max-w-6xl mx-auto px-6 py-16 md:py-24">
      <h2 class="text-3xl md:text-4xl font-bold text-center text-zinc-100 mb-12">
        Request an Enterprise Extension
      </h2>

      <div class="max-w-2xl mx-auto">
        <form class="space-y-6" onSubmit={handleSubmit}>
          <div>
            <label for="name" class="block text-sm font-medium text-zinc-300">
              Name
            </label>
            <input
              type="text"
              id="name"
              name="name"
              required
              class="w-full py-3 px-4 bg-zinc-900 border border-zinc-800/80 focus:outline-none focus:ring-2 focus:ring-cyan-400/20 rounded-lg"
            />
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-zinc-300">
              Work Email
            </label>
            <input
              type="email"
              id="email"
              name="email"
              required
              class="w-full py-3 px-4 bg-zinc-900 border border-zinc-800/80 focus:outline-none focus:ring-2 focus:ring-cyan-400/20 rounded-lg"
            />
          </div>

          <div>
            <label for="subject" class="block text-sm font-medium text-zinc-300">
              Subject
            </label>
            <select
              id="subject"
              name="subject"
              required
              class="w-full py-3 px-4 bg-zinc-900 border border-zinc-800/80 focus:outline-none focus:ring-2 focus:ring-cyan-400/20 rounded-lg"
            >
              <option value="Quota Extension Request">Quota Extension Request</option>
              <option value="General Inquiry">General Inquiry</option>
              <option value="Enterprise License">Enterprise License</option>
            </select>
          </div>

          <div>
            <label for="message" class="block text-sm font-medium text-zinc-300">
              Message
            </label>
            <textarea
              id="message"
              name="message"
              required
              rows="5"
              class="w-full py-3 px-4 bg-zinc-900 border border-zinc-800/80 focus:outline-none focus:ring-2 focus:ring-cyan-400/20 rounded-lg"
            />
          </div>

          <button
            type="submit"
            disabled={loading()}
            class="w-full py-3 px-4 bg-cyan-500 hover:bg-cyan-400 text-zinc-950 font-medium rounded-lg transition-colors disabled:opacity-50"
          >
            {loading() ? 'Sending...' : 'Submit Request'}
          </button>

          <Show when={submitted()}>
            <p class="text-green-400 text-sm text-center">
              Your request has been dispatched. Our team will adjust your quota shortly.
            </p>
          </Show>
        </form>
      </div>
    </section>
  );
};