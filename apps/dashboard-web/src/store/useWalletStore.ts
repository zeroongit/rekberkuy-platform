import { create } from 'zustand';
import { RekberPayWallet } from '@/types';

interface WalletState {
  wallet: RekberPayWallet | null;
  isLoading: boolean;
  setWallet: (wallet: RekberPayWallet) => void;
  updateBalance: (balance: number, lockedBalance: number) => void;
  setLoading: (isLoading: boolean) => void;
}

export const useWalletStore = create<WalletState>((set) => ({
  wallet: {
    userId: 'mock-user-1',
    balance: 2500000, // Rp 2.500.000 default mock
    lockedBalance: 500000, // Rp 500.000 escrow locked
    currency: 'IDR',
    updatedAt: new Date().toISOString(),
  },
  isLoading: false,
  setWallet: (wallet) => set({ wallet }),
  updateBalance: (balance, lockedBalance) =>
    set((state) => ({
      wallet: state.wallet ? { ...state.wallet, balance, lockedBalance, updatedAt: new Date().toISOString() } : null,
    })),
  setLoading: (isLoading) => set({ isLoading }),
}));
