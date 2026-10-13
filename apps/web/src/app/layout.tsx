import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: { default: "Aravindh Varadharaj · Python & Go Developer", template: "%s · Aravindh Varadharaj" },
  description: "Aravindh Varadharaj's Python, Go and Next.js developer portfolio: open learning, historical famine analysis, animal welfare, disaster impact, nuclear civil safety and wildlife-aware network efficiency.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <a className="skip-link" href="#main">Skip to content</a>
        <header className="site-header">
          <a className="brand" href="/"><span>Aravindh Varadharaj</span><b>Developer Portfolio</b></a>
          <nav aria-label="Primary">
            <a href="/#systems">Projects</a>
            <a href="/#stack">Skills</a>
            <a href="/#profile">Experience</a>
            <a href="/resume.pdf">Resume</a>
            <a href="mailto:aravindhvaradha@gmail.com">Contact</a>
          </nav>
        </header>
        {children}
        <footer><span>Portfolio built with Next.js, React, TypeScript, Python and Go.</span><span>Aravindh Varadharaj · 2026</span></footer>
      </body>
    </html>
  );
}
