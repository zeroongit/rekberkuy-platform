import React from 'react';

export interface Column<T> {
  header: string;
  accessorKey?: keyof T | ((row: T) => React.ReactNode);
  cell?: (row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  data: T[];
  columns: Column<T>[];
  keyExtractor: (row: T) => string;
  emptyMessage?: string;
}

export function DataTable<T>({
  data,
  columns,
  keyExtractor,
  emptyMessage = 'Tidak ada data.',
}: DataTableProps<T>) {
  if (!data || data.length === 0) {
    return (
      <div className="p-8 text-center text-zinc-500 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl">
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl shadow-sm">
      <table className="w-full text-left border-collapse">
        <thead>
          <tr className="border-b border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-900/50 text-xs font-semibold uppercase text-zinc-500">
            {columns.map((col, idx) => (
              <th key={idx} className="px-6 py-4">
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-zinc-200 dark:divide-zinc-800 text-sm">
          {data.map((row) => (
            <tr key={keyExtractor(row)} className="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition">
              {columns.map((col, idx) => {
                let content: React.ReactNode;
                if (col.cell) {
                  content = col.cell(row);
                } else if (typeof col.accessorKey === 'function') {
                  content = col.accessorKey(row);
                } else if (col.accessorKey) {
                  content = String(row[col.accessorKey] ?? '');
                } else {
                  content = '';
                }
                return (
                  <td key={idx} className="px-6 py-4 text-zinc-800 dark:text-zinc-200">
                    {content}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
