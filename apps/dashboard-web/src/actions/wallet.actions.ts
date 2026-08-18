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
    // In production, call core-service backend API using ctx.authHeaders & ctx.token
    // Example:
    // const res = await fetch(`${process.env.CORE_SERVICE_URL}/api/v1/wallet/topup`, {
    //   method: 'POST',
    //   headers: {
    //     'Content-Type': 'application/json',
    //     ...ctx.authHeaders,
    //   },
    //   body: JSON.stringify(data),
    // });
    // if (!res.ok) throw new Error('Gagal memproses top-up wallet');
    // return await res.json();

    // Mock successful response for demonstration & prototyping
    console.log('[Topup Action] Processing topup with token:', ctx.token ? 'Authenticated' : 'Unauthenticated');
    console.log('[Topup Action] Validated & Sanitized Data:', data);

    return {
      paymentUrl: 'https://app.sandbox.midtrans.com/snap/v2/vtweb/mock-snap-token',
      snapToken: 'mock-snap-token-123456',
      orderId: 'TOPUP-' + Date.now(),
    };
  },
  { requireAuth: true, requireCsrf: true }
);
