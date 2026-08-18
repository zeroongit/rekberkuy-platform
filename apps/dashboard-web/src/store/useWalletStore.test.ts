import { useWalletStore } from './useWalletStore';

describe('useWalletStore Zustand Store', () => {
  beforeEach(() => {
    useWalletStore.setState({
      wallet: {
        userId: 'test-user',
        balance: 1000000,
        lockedBalance: 200000,
        currency: 'IDR',
        updatedAt: new Date().toISOString(),
      },
      isLoading: false,
    });
  });

  it('should update balance correctly', () => {
    const { updateBalance } = useWalletStore.getState();
    updateBalance(1500000, 300000);

    const updatedWallet = useWalletStore.getState().wallet;
    expect(updatedWallet?.balance).toBe(1500000);
    expect(updatedWallet?.lockedBalance).toBe(300000);
  });
});
