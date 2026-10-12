import { getAuthHeaders } from '@/lib/security/jwt';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

async function apiFetch<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const authHeaders = await getAuthHeaders();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...authHeaders,
    ...options.headers,
  };

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({ error: 'Terjadi kesalahan pada server' }));
    throw new Error(errorBody.error || errorBody.message || `HTTP error! status: ${res.status}`);
  }

  return res.json() as Promise<T>;
}

export interface RequestOptions extends RequestInit {
  headers?: Record<string, string>;
}

export const apiService = {
  // Auth
  register: (data: unknown) => apiFetch('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: unknown) => apiFetch('/auth/login', { method: 'POST', body: JSON.stringify(data) }),
  tokenTest: (data: unknown) => apiFetch('/users/token-test', { method: 'POST', body: JSON.stringify(data) }),

  // KYC & Vendor
  submitKyc: (data: unknown) => apiFetch('/kyc/submit', { method: 'POST', body: JSON.stringify(data) }),
  registerVendor: (data: unknown) => apiFetch('/vendors/register', { method: 'POST', body: JSON.stringify(data) }),

  // Transactions
  listTransactions: () => apiFetch('/transactions'),
  getTransactionDetail: (id: string) => apiFetch(`/transactions/${id}`),
  lockGoods: (data: unknown) => apiFetch('/transactions/goods/lock', { method: 'POST', body: JSON.stringify(data) }),
  releaseGoods: (data: unknown) => apiFetch('/transactions/goods/release', { method: 'POST', body: JSON.stringify(data) }),
  lockServices: (data: unknown) => apiFetch('/transactions/services/lock', { method: 'POST', body: JSON.stringify(data) }),
  releaseMilestone: (data: unknown) => apiFetch('/transactions/services/release-milestone', { method: 'POST', body: JSON.stringify(data) }),
  lockEvents: (data: unknown) => apiFetch('/transactions/events/lock', { method: 'POST', body: JSON.stringify(data) }),
  submitVendorInvoice: (id: string, data: unknown) => apiFetch(`/transactions/events/${id}/vendor-invoices`, { method: 'POST', body: JSON.stringify(data) }),
  releaseEventMilestone: (data: unknown) => apiFetch('/transactions/events/release-milestone', { method: 'POST', body: JSON.stringify(data) }),
  processEventVendorPayout: (data: unknown) => apiFetch('/transactions/events/release-vendors', { method: 'POST', body: JSON.stringify(data) }),
  markEventDisbursement: (id: string) => apiFetch(`/transactions/events/payouts/${id}/disburse`, { method: 'POST' }),

  // Catalog & Marketplace
  getCategories: () => apiFetch('/categories'),
  listMarketplaceVendors: () => apiFetch('/vendors'),

  // Wallet
  topupWallet: (data: unknown) => apiFetch('/wallets/topup', { method: 'POST', body: JSON.stringify(data) }),
  getWalletBalance: () => apiFetch('/wallets/me'),
  getWalletHistory: () => apiFetch('/wallets/me/transactions'),
  requestWithdrawal: (data: unknown) => apiFetch('/wallets/withdraw', { method: 'POST', body: JSON.stringify(data) }),
  listWithdrawals: () => apiFetch('/wallets/withdrawals'),

  // Reviews
  createReview: (data: unknown) => apiFetch('/reviews', { method: 'POST', body: JSON.stringify(data) }),
  getReviewsForUser: (userId: string) => apiFetch(`/users/${userId}/reviews`),

  // Disputes
  openDispute: (data: unknown) => apiFetch('/disputes', { method: 'POST', body: JSON.stringify(data) }),
  getDispute: (id: string) => apiFetch(`/disputes/${id}`),
  acknowledgeDispute: (id: string) => apiFetch(`/disputes/${id}/acknowledge`, { method: 'POST' }),
  resolveDispute: (id: string, data: unknown) => apiFetch(`/disputes/${id}/resolve`, { method: 'POST', body: JSON.stringify(data) }),

  // Admin
  adminGetPendingKYCs: () => apiFetch('/admin/kyc/pending'),
  adminGetKYCDetail: (id: string) => apiFetch(`/admin/kyc/${id}`),
  adminReviewKYC: (id: string, data: unknown) => apiFetch(`/admin/kyc/${id}/review`, { method: 'POST', body: JSON.stringify(data) }),
  adminListPendingWithdrawals: () => apiFetch('/admin/withdrawals/pending'),
  adminDisburseWithdrawal: (id: string) => apiFetch(`/admin/withdrawals/${id}/disburse`, { method: 'POST' }),
};
