import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';

const features = [
  {
    icon: '~>',
    title: 'Migrate & Undo',
    desc: 'Apply versioned migrations forward or roll them back. No paid tier required for undo.',
  },
  {
    icon: '{}',
    title: 'Dry Run',
    desc: 'Preview exactly what SQL will run before committing. Catch issues before they hit production.',
  },
  {
    icon: '#>',
    title: 'Cherry-Pick & Skip',
    desc: 'Apply specific migrations out of order or skip problematic ones. Full control over your pipeline.',
  },
  {
    icon: '<>',
    title: 'Schema Diff',
    desc: 'Compare database schemas against snapshots or other databases. See exactly what changed.',
  },
  {
    icon: '!?',
    title: 'SQL Linting',
    desc: 'Catch dangerous patterns like DROP TABLE, missing indexes, and naming violations before they ship.',
  },
  {
    icon: '[]',
    title: 'Squash Migrations',
    desc: 'Consolidate dozens of migration files into a single clean baseline. Keep your history tidy.',
  },
  {
    icon: '::',
    title: 'Web Dashboard',
    desc: 'Embedded UI for visualizing migration state, schema diffs, and lint results. Zero extra setup.',
  },
  {
    icon: '>>',
    title: 'Single Binary',
    desc: 'One binary, zero runtime dependencies. Install in seconds via Homebrew, curl, or Docker.',
  },
  {
    icon: 'DB',
    title: 'Multi-Database',
    desc: 'First-class support for PostgreSQL, MySQL, and SQLite. Same CLI, same workflow, any database.',
  },
];

function Hero() {
  return (
    <section className="hero-section">
      <h1 className="hero-title">
        Database migrations,
        <br />
        <span className="gradient-text">fully open source</span>
      </h1>
      <p className="hero-subtitle">
        Everything Flyway paywalls — undo, dry-run, cherry-pick, schema diff, linting — is free.
        Single binary, zero dependencies, instant setup.
      </p>
      <div className="hero-actions">
        <Link className="hero-btn hero-btn-primary" to="/docs/getting-started">
          Get Started
        </Link>
        <Link className="hero-btn hero-btn-secondary" to="https://github.com/dunkinfrunkin/drift">
          View on GitHub
        </Link>
      </div>
      <div className="install-snippet">
        <code>brew install dunkinfrunkin/tap/drift</code>
      </div>
    </section>
  );
}

function Features() {
  return (
    <section className="features-section">
      <div className="features-section-title">Features</div>
      <h2 className="features-section-heading">Everything you need, nothing you don't</h2>
      <p className="features-section-desc">
        A complete migration toolkit that works out of the box. No license keys, no feature gates, no surprises.
      </p>
      <div className="features-grid">
        {features.map((f) => (
          <div key={f.title} className="feature-card">
            <div className="feature-icon">
              <span style={{ fontFamily: 'var(--ifm-font-family-monospace)', fontWeight: 700, fontSize: '0.85rem', color: 'var(--feature-icon-color)' }}>
                {f.icon}
              </span>
            </div>
            <h3>{f.title}</h3>
            <p>{f.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

function Databases() {
  return (
    <section className="db-section">
      <div className="features-section-title">Database Support</div>
      <h2 className="features-section-heading">Works with your stack</h2>
      <div className="db-badges">
        <span className="db-badge">PostgreSQL</span>
        <span className="db-badge">MySQL</span>
        <span className="db-badge">SQLite</span>
      </div>
    </section>
  );
}

function CTA() {
  return (
    <section className="cta-section">
      <h2 className="features-section-heading">Ready to migrate?</h2>
      <p className="features-section-desc" style={{ marginBottom: '2rem' }}>
        Get up and running in under a minute. No account needed.
      </p>
      <div className="hero-actions">
        <Link className="hero-btn hero-btn-primary" to="/docs/getting-started">
          Read the Docs
        </Link>
        <Link className="hero-btn hero-btn-secondary" to="/docs/installation">
          Install Drift
        </Link>
      </div>
    </section>
  );
}

export default function Home(): React.JSX.Element {
  return (
    <Layout
      title="Database Migrations, Fully Open Source"
      description="Fast, open-source database migration tool with undo, dry-run, diff, and linting. Single binary, zero dependencies."
    >
      <Hero />
      <Features />
      <Databases />
      <CTA />
    </Layout>
  );
}
