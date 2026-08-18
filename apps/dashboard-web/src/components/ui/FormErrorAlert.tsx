import React from 'react';

interface FormErrorAlertProps {
  message?: string | null;
  errors?: Record<string, string[]> | null;
  className?: string;
}

export function FormErrorAlert({ message, errors, className = '' }: FormErrorAlertProps) {
  if (!message && (!errors || Object.keys(errors).length === 0)) {
    return null;
  }

  return (
    <div
      role="alert"
      className={`p-4 mb-4 bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 rounded-xl text-red-800 dark:text-red-200 shadow-sm ${className}`}
    >
      <div className="flex items-start space-x-3">
        <svg
          className="w-5 h-5 text-red-600 dark:text-red-400 mt-0.5 shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <div className="flex-1 text-sm">
          {message && <p className="font-semibold">{message}</p>}
          {errors && Object.keys(errors).length > 0 && (
            <ul className="mt-2 list-disc list-inside space-y-1 text-xs opacity-90">
              {Object.entries(errors).map(([field, errs]) => (
                <li key={field}>
                  <span className="font-medium capitalize">{field}:</span> {errs.join(', ')}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
