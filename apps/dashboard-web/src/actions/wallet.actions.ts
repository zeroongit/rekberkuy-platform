'use server';

import { createServerAction } from '@/lib/action-client';
import { topupWalletSchema, TopupWalletInput } from '@/lib/validations/wallet.schema';

export interface TopupResult {
  paymentUrl: string;
  snapToken: string;
  orderId: string;
}

/**
 * Secure Server Action for Wallet Top-up.
 * Protected by CSRF token verification, JWT Auth extraction, Zod validation, and input sanitization.
 */
export const topupWalletAction = createServerAction(
  topupWalletSchema,
  async (data: TopupWalletInput, ctx) => {
    const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

    const res = await fetch(`${apiBase}/wallets/topup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...ctx.authHeaders,
      },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'Gagal memproses top-up wallet' }));
      throw new Error(errBody.error || 'Gagal memproses top-up wallet');
    }

    return await res.json();
  },
  { requireAuth: true, requireCsrf: true }
);
