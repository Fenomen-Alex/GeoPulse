import type { Component } from 'solid-js';
import { LandingHeader } from '../components/landing/LandingHeader';
import { Hero } from '../components/landing/Hero';
import { ValueProps } from '../components/landing/ValueProps';
import { TechStack } from '../components/landing/TechStack';
import { QuotaPricing } from '../components/landing/QuotaPricing';
import { ContactForm } from '../components/landing/ContactForm';
import { LandingFooter } from '../components/landing/LandingFooter';

export const LandingView: Component = () => {
  return (
    <div class="min-h-screen bg-zinc-950 text-zinc-100">
      <LandingHeader />
      <Hero />
      <div id="features">
        <ValueProps />
      </div>
      <div id="tech-stack">
        <TechStack />
      </div>
      <div id="quotas">
        <QuotaPricing />
      </div>
      <div id="contact">
        <ContactForm />
      </div>
      <LandingFooter />
    </div>
  );
};