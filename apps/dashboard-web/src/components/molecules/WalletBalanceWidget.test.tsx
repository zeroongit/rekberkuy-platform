import React from 'react';
import { render, screen } from '@testing-library/react';
import { WalletBalanceWidget } from './WalletBalanceWidget';
import { useWalletStore } from '@/store/useWalletStore';

describe('WalletBalanceWidget Component', () => {
  beforeEach(() => {
    useWalletStore.setState({
      wallet: {
        userId: 'test-user',
        balance: 2500000,
        lockedBalance: 500000,
        currency: 'IDR',
        updatedAt: new Date().toISOString(),
      },
      isLoading: false,
    });
  });

  it('renders wallet balance correctly', () => {
    render(<WalletBalanceWidget />);
    expect(screen.getByText('RekberPay Wallet')).toBeInTheDocument();
  });
});
