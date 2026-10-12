'use server';

import { createServerAction } from '@/lib/action-client';
import { kycSubmissionSchema, KYCSubmissionInput } from '@/lib/validations/auth.schema';

/**
 * Secure Server Action for KYC Submission.
 */
export const submitKycAction = createServerAction(
  kycSubmissionSchema,
  async (data: KYCSubmissionInput, ctx) => {
    const apiBase = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

    const res = await fetch(`${apiBase}/kyc/submit`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...ctx.authHeaders,
      },
      body: JSON.stringify(data),
    });

    if (!res.ok) {
      const errBody = await res.json().catch(() => ({ error: 'Gagal mengirim KYC' }));
      throw new Error(errBody.error || 'Gagal mengirim pengajuan KYC');
    }

    return await res.json();
  },
  { requireAuth: true, requireCsrf: true }
);
