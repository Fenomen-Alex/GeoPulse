import { type Component, createSignal } from 'solid-js';
import { LandingHeader } from '../components/landing/LandingHeader';
import { Hero } from '../components/landing/Hero';
import { ValueProps } from '../components/landing/ValueProps';
import { TechStack } from '../components/landing/TechStack';
import { QuotaPricing } from '../components/landing/QuotaPricing';
import { ContactForm } from '../components/landing/ContactForm';
import { LandingFooter } from '../components/landing/LandingFooter';
import { AuthChoiceModal } from '../components/landing/AuthChoiceModal';

export const LandingView: Component = () => {
  const [authModalOpen, setAuthModalOpen] = createSignal(false);

  return (
    <div class="min-h-screen bg-zinc-950 text-zinc-100">
      <LandingHeader onOpenAuth={() => setAuthModalOpen(true)} />
      <Hero onOpenAuth={() => setAuthModalOpen(true)} />
      <div id="features">
        <ValueProps />
      </div>
      <div id="tech-stack">
        <TechStack />
      </div>
      <div id="quotas">
        <QuotaPricing onOpenAuth={() => setAuthModalOpen(true)} />
      </div>
      <div id="contact">
        <ContactForm />
      </div>
      <LandingFooter />
      <AuthChoiceModal open={authModalOpen()} onClose={() => setAuthModalOpen(false)} />
    </div>
  );
};
