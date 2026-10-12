import { topupWalletSchema } from './wallet.schema';

describe('TopupWallet Zod Schema Validation', () => {
  it('should validate valid top-up input successfully', () => {
    const validData = {
      amount: 50000,
      paymentMethod: 'VIRTUAL_ACCOUNT' as const,
    };
    const result = topupWalletSchema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('should fail when amount is below minimum (10.000 IDR)', () => {
    const invalidData = {
      amount: 5000,
      paymentMethod: 'QRIS' as const,
    };
    const result = topupWalletSchema.safeParse(invalidData);
    expect(result.success).toBe(false);
  });

  it('should fail when payment method is invalid', () => {
    const invalidData = {
      amount: 50000,
      paymentMethod: 'INVALID_METHOD',
    };
    const result = topupWalletSchema.safeParse(invalidData);
    expect(result.success).toBe(false);
  });
});
