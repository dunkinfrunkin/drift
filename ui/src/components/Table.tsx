import { clsx } from 'clsx';

export function Table({ className, children }: React.HTMLAttributes<HTMLTableElement>) {
  return (
    <div className="overflow-x-auto">
      <table className={clsx('w-full text-sm', className)}>{children}</table>
    </div>
  );
}

export function Thead({ children }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <thead className="text-xs text-slate-500 uppercase tracking-wider">{children}</thead>;
}

export function Th({ className, children }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return <th className={clsx('px-4 py-3 text-left font-medium', className)}>{children}</th>;
}

export function Td({ className, children }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return <td className={clsx('px-4 py-3 border-t border-slate-800', className)}>{children}</td>;
}
