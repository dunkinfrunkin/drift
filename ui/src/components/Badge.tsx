import { clsx } from 'clsx';

const variants: Record<string, string> = {
  success: 'bg-green-500/15 text-green-400 border-green-500/25',
  warning: 'bg-yellow-500/15 text-yellow-400 border-yellow-500/25',
  error: 'bg-red-500/15 text-red-400 border-red-500/25',
  info: 'bg-blue-500/15 text-blue-400 border-blue-500/25',
  muted: 'bg-slate-500/15 text-slate-400 border-slate-500/25',
};

interface BadgeProps {
  variant?: keyof typeof variants;
  children: React.ReactNode;
  className?: string;
}

export function Badge({ variant = 'muted', children, className }: BadgeProps) {
  return (
    <span className={clsx(
      'inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full border',
      variants[variant],
      className,
    )}>
      {children}
    </span>
  );
}

export function stateBadgeVariant(state: string): keyof typeof variants {
  switch (state) {
    case 'Applied': return 'success';
    case 'Pending': return 'warning';
    case 'Failed': return 'error';
    case 'Missing': return 'error';
    default: return 'muted';
  }
}
