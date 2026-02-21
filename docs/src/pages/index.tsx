import React, { useEffect, useState } from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import {
  ArrowLeftRight,
  Beaker,
  ChevronRight,
  CircleCheck,
  CircleX,
  DatabaseZap,
  Eye,
  GitBranch,
  LayoutDashboard,
  Package,
  ScissorsLineCut,
  ShieldCheck,
  Shrink,
  X,
} from 'lucide-react';

// ─── Animated Terminal ───

const terminalLines: { text: string; delay: number; type: string }[] = [
  { text: '$ drift migrate --dry-run', delay: 0, type: 'command' },
  { text: '', delay: 400, type: 'blank' },
  { text: '  Database: postgresql://localhost/myapp', delay: 600, type: 'dim' },
  { text: '  Table:    drift_schema_history', delay: 750, type: 'dim' },
  { text: '', delay: 900, type: 'blank' },
  { text: '  DRY RUN — no changes will be applied', delay: 1100, type: 'warn' },
  { text: '', delay: 1300, type: 'blank' },
  { text: '  V003  create_orders', delay: 1500, type: 'pending' },
  { text: '  V004  add_order_indexes', delay: 1700, type: 'pending' },
  { text: '', delay: 1900, type: 'blank' },
  { text: '  2 pending migrations would be applied.', delay: 2100, type: 'info' },
  { text: '', delay: 2500, type: 'blank' },
  { text: '$ drift migrate', delay: 2800, type: 'command' },
  { text: '', delay: 3100, type: 'blank' },
  { text: '  V003  create_orders ............. ✓  38ms', delay: 3400, type: 'success' },
  { text: '  V004  add_order_indexes ......... ✓  12ms', delay: 3900, type: 'success' },
  { text: '', delay: 4200, type: 'blank' },
  { text: '  2 migrations applied successfully.', delay: 4400, type: 'done' },
];

function TerminalDemo() {
  const [visibleCount, setVisibleCount] = useState(0);

  useEffect(() => {
    const timers: ReturnType<typeof setTimeout>[] = [];
    terminalLines.forEach((line, i) => {
      timers.push(setTimeout(() => setVisibleCount(i + 1), line.delay));
    });
    return () => timers.forEach(clearTimeout);
  }, []);

  return (
    <div className="terminal-window">
      <div className="terminal-header">
        <div className="terminal-dot terminal-dot-red" />
        <div className="terminal-dot terminal-dot-yellow" />
        <div className="terminal-dot terminal-dot-green" />
        <span className="terminal-title">drift</span>
      </div>
      <div className="terminal-body">
        {terminalLines.slice(0, visibleCount).map((line, i) => (
          <div key={i} className={`terminal-line terminal-${line.type}`}>
            {line.text}
          </div>
        ))}
        {visibleCount < terminalLines.length && (
          <span className="terminal-cursor">▋</span>
        )}
      </div>
    </div>
  );
}

// ─── Features ───

const features = [
  {
    icon: <GitBranch size={20} />,
    title: 'Migrate & Undo',
    desc: 'Apply migrations forward or roll them back. Undo is free — no paid tier.',
  },
  {
    icon: <Eye size={20} />,
    title: 'Dry Run',
    desc: 'Preview exactly what SQL will execute before committing to anything.',
  },
  {
    icon: <ScissorsLineCut size={20} />,
    title: 'Cherry-Pick & Skip',
    desc: 'Apply specific migrations out of order or skip problematic ones.',
  },
  {
    icon: <ArrowLeftRight size={20} />,
    title: 'Schema Diff',
    desc: 'Compare schemas against snapshots. See exactly what changed and why.',
  },
  {
    icon: <ShieldCheck size={20} />,
    title: 'SQL Linting',
    desc: 'Catch DROP TABLE, naming violations, and dangerous patterns before deploy.',
  },
  {
    icon: <Shrink size={20} />,
    title: 'Squash',
    desc: 'Consolidate migration files into a clean baseline. Keep history tidy.',
  },
  {
    icon: <LayoutDashboard size={20} />,
    title: 'Web Dashboard',
    desc: 'Embedded UI for migration state, diffs, and lint results. Zero setup.',
  },
  {
    icon: <Package size={20} />,
    title: 'Single Binary',
    desc: 'One binary, zero dependencies. Install in seconds via Homebrew or curl.',
  },
  {
    icon: <DatabaseZap size={20} />,
    title: 'Multi-Database',
    desc: 'PostgreSQL, MySQL, and SQLite. Same CLI, same workflow.',
  },
];

// ─── Flyway Comparison ───

const comparisonRows = [
  { feature: 'Undo migrations', flyway: 'Teams plan', drift: 'Free' },
  { feature: 'Dry-run mode', flyway: 'Teams plan', drift: 'Free' },
  { feature: 'Cherry-pick', flyway: 'Teams plan', drift: 'Free' },
  { feature: 'Schema diff', flyway: 'Not available', drift: 'Free' },
  { feature: 'SQL linting', flyway: 'Not available', drift: 'Free' },
  { feature: 'Migration squash', flyway: 'Not available', drift: 'Free' },
  { feature: 'Web dashboard', flyway: 'Not available', drift: 'Free' },
];

// ─── Sections ───

function Hero() {
  const [copied, setCopied] = useState(false);
  const installCmd = 'brew install dunkinfrunkin/tap/drift';

  function handleCopy() {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <section className="hero-section">
      <div className="hero-grid">
        <div className="hero-content">
          <div className="hero-badge">Open source database migrations</div>
          <h1 className="hero-title">
            Migrate databases
            <br />
            <span className="gradient-text">without the paywall</span>
          </h1>
          <p className="hero-subtitle">
            Everything Flyway locks behind a paid plan — undo, dry-run,
            cherry-pick, schema diff, SQL linting — is free. Single binary.
            Zero dependencies.
          </p>
          <div className="hero-actions">
            <Link className="hero-btn hero-btn-primary" to="/docs/getting-started">
              Get Started
              <ChevronRight size={16} />
            </Link>
            <Link className="hero-btn hero-btn-secondary" to="https://github.com/dunkinfrunkin/drift">
              View on GitHub
            </Link>
          </div>
          <div className="install-snippet" onClick={handleCopy} role="button" tabIndex={0}>
            <code>
              <span className="install-prompt">$</span> {installCmd}
            </code>
            <span className="install-copy">{copied ? 'Copied!' : 'Copy'}</span>
          </div>
        </div>
        <div className="hero-terminal">
          <TerminalDemo />
        </div>
      </div>
    </section>
  );
}

function Comparison() {
  return (
    <section className="comparison-section">
      <div className="comparison-inner">
        <div className="section-label">Why Drift?</div>
        <h2 className="section-heading">The Flyway tax, eliminated</h2>
        <p className="section-desc">
          Flyway locks critical migration features behind expensive plans.
          Drift gives you everything — for free, forever.
        </p>
        <div className="comparison-table-wrap">
          <table className="comparison-table">
            <thead>
              <tr>
                <th>Feature</th>
                <th>Flyway Community</th>
                <th className="comparison-drift-col">Drift</th>
              </tr>
            </thead>
            <tbody>
              {comparisonRows.map((row) => (
                <tr key={row.feature}>
                  <td className="comparison-feature">{row.feature}</td>
                  <td className="comparison-flyway">
                    <CircleX size={16} className="comparison-x" />
                    <span>{row.flyway}</span>
                  </td>
                  <td className="comparison-drift">
                    <CircleCheck size={16} className="comparison-check" />
                    <span>{row.drift}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}

function Features() {
  return (
    <section className="features-section">
      <div className="section-label">Features</div>
      <h2 className="section-heading">Everything you need, nothing you don't</h2>
      <p className="section-desc">
        A complete migration toolkit that works out of the box.
        No license keys, no feature gates, no surprises.
      </p>
      <div className="features-grid">
        {features.map((f) => (
          <div key={f.title} className="feature-card">
            <div className="feature-icon">{f.icon}</div>
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
      <div className="section-label">Database Support</div>
      <h2 className="section-heading">Works with your stack</h2>
      <div className="db-cards">
        <div className="db-card">
          <div className="db-card-name">PostgreSQL</div>
          <div className="db-card-detail">pgx/v5 driver, advisory locks, transactional DDL</div>
        </div>
        <div className="db-card">
          <div className="db-card-name">MySQL</div>
          <div className="db-card-detail">go-sql-driver, GET_LOCK concurrency, information_schema</div>
        </div>
        <div className="db-card">
          <div className="db-card-name">SQLite</div>
          <div className="db-card-detail">Pure Go driver, zero CGO, WAL mode, single-file DB</div>
        </div>
      </div>
    </section>
  );
}

function CTA() {
  return (
    <section className="cta-section">
      <h2 className="section-heading">Ready to migrate?</h2>
      <p className="section-desc" style={{ marginBottom: '2rem' }}>
        Get up and running in under a minute. No account required.
      </p>
      <div className="hero-actions" style={{ justifyContent: 'center' }}>
        <Link className="hero-btn hero-btn-primary" to="/docs/getting-started">
          Read the Docs
          <ChevronRight size={16} />
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
      <Comparison />
      <Features />
      <Databases />
      <CTA />
    </Layout>
  );
}
