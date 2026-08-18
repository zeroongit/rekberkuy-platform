import { z } from 'zod';

export const topupWalletSchema = z.object({
  amount: z.number().positive('Jumlah top-up harus lebih besar dari 0').min(10000, 'Minimum top-up adalah Rp 10.000'),
  paymentMethod: z.enum(['BANK_TRANSFER', 'QRIS', 'VIRTUAL_ACCOUNT'], {
    error: 'Metode pembayaran tidak valid',
  }),
});

export const withdrawWalletSchema = z.object({
  amount: z.number().positive('Jumlah penarikan harus lebih besar dari 0').min(50000, 'Minimum penarikan adalah Rp 50.000'),
  bankName: z.string().min(2, 'Nama bank wajib diisi'),
  bankAccountNumber: z.string().min(5, 'Nomor rekening bank wajib diisi'),
  bankAccountHolder: z.string().min(2, 'Nama pemilik rekening wajib diisi'),
});

export type TopupWalletInput = z.infer<typeof topupWalletSchema>;
export type WithdrawWalletInput = z.infer<typeof withdrawWalletSchema>;
