'use server';

import { createServerAction } from '@/lib/action-client';
import { createDisputeSchema, CreateDisputeInput } from '@/lib/validations/transaction.schema';

/**
 * Secure Server Action for creating a Dispute.
 */
export const createDisputeAction = createServerAction(
  createDisputeSchema,
  async (data: CreateDisputeInput, ctx) => {
    const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

    const res = await fetch(`${apiBase}/disputes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...ctx.authHeaders,
      },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'Gagal mengajukan sengketa' }));
      throw new Error(errBody.error || 'Gagal mengajukan sengketa');
    }

    return await res.json();
  },
  { requireAuth: true, requireCsrf: true }
);
