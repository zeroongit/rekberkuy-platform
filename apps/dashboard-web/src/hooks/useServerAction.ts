'use client';

import { useState, useTransition } from 'react';
import { z } from 'zod';
import { sanitizeObject } from '@/lib/security/sanitize';

export interface ActionResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  validationErrors?: Record<string, string[]>;
}

export function useServerAction<TSchema extends z.ZodTypeAny, TResult>(
  schema: TSchema,
  actionFn: (data: z.infer<TSchema>) => Promise<ActionResponse<TResult>>
) {
  const [isPending, startTransition] = useTransition();
  const [data, setData] = useState<TResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]> | null>(null);
  const [success, setSuccess] = useState<boolean>(false);

  const execute = async (rawInput: z.infer<TSchema>) => {
    setError(null);
    setValidationErrors(null);
    setSuccess(false);

    // 1. Sanitize input
    const sanitizedInput = sanitizeObject(rawInput);

    // 2. Client-side Zod validation
    const result = schema.safeParse(sanitizedInput);
    if (!result.success) {
      const formattedErrors = result.error.flatten().fieldErrors as Record<string, string[]>;
      setValidationErrors(formattedErrors);
      setError('Validasi gagal. Periksa kembali input Anda.');
      return {
        success: false,
        validationErrors: formattedErrors,
        error: 'Validasi gagal',
      };
    }

    return new Promise<ActionResponse<TResult>>((resolve) => {
      startTransition(async () => {
        try {
          // 3. Call server action
          const response = await actionFn(result.data);

          if (response.success) {
            setData(response.data ?? null);
            setSuccess(true);
            setError(null);
            setValidationErrors(null);
          } else {
            setError(response.error || 'Terjadi kesalahan pada server');
            if (response.validationErrors) {
              setValidationErrors(response.validationErrors);
            }
          }
          resolve(response);
        } catch (err: unknown) {
          const errorMessage = err instanceof Error ? err.message : 'Terjadi kesalahan sistem';
          setError(errorMessage);
          resolve({
            success: false,
            error: errorMessage,
          });
        }
      });
    });
  };

  return {
    execute,
    isPending,
    data,
    error,
    validationErrors,
    success,
  };
}
