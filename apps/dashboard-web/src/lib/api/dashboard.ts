import { getServerJwtToken } from '@/lib/security/jwt.server';

export async function getDashboardStats() {
  const apiBase = process.env.API_INTERNAL_URL || 'http://localhost:8080/api/v1';
  try {
    const token = await getServerJwtToken();
    const headers: Record<string, string> = token ? { Authorization: `Bearer ${token}` } : {};

    let walletBalance = 0;
    try {
      const walletRes = await fetch(`${apiBase}/wallets/me`, { headers, cache: 'no-store' });
      if (walletRes.ok) {
        const walletData = await walletRes.json();
        walletBalance = walletData.balance || walletData.wallet?.balance || 0;
      }
    } catch {
      // ignore
    }

    let activeTransactions = 0;
    let completedTransactions = 0;
    try {
      const txRes = await fetch(`${apiBase}/transactions`, { headers, cache: 'no-store' });
      if (txRes.ok) {
        const txData = await txRes.json();
        const txs = Array.isArray(txData) ? txData : (txData.transactions || txData.data || []);
        activeTransactions = txs.filter((t: { status: string }) => t.status !== 'RELEASED' && t.status !== 'REFUNDED').length;
        completedTransactions = txs.filter((t: { status: string }) => t.status === 'RELEASED').length;
      }
    } catch {
      // ignore
    }

    return {
      walletBalance,
      activeTransactions,
      completedTransactions,
    };
  } catch {
    return {
      walletBalance: 0,
      activeTransactions: 0,
      completedTransactions: 0,
    };
  }
}
