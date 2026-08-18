import React from 'react';

interface FieldErrorProps {
  message?: string;
  className?: string;
}

export function FieldError({ message, className = '' }: FieldErrorProps) {
  if (!message) return null;

  return (
    <p className={`mt-1 text-xs font-medium text-red-600 dark:text-red-400 flex items-center space-x-1 ${className}`}>
      <span>⚠️</span>
      <span>{message}</span>
    </p>
  );
}
