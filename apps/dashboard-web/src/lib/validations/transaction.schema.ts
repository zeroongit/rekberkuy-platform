import { z } from 'zod';

export const transactionTypeEnum = z.enum(['GOODS', 'SERVICES', 'EVENTS']);

export const milestoneSchema = z.object({
  title: z.string().min(3, 'Judul milestone minimal 3 karakter'),
  amount: z.number().positive('Nominal milestone harus lebih besar dari 0'),
  dueDate: z.string().optional(),
});

export const eventVendorPayoutSchema = z.object({
  vendorUserId: z.string().uuid().optional(),
  vendorName: z.string().min(2, 'Nama vendor wajib diisi'),
  amount: z.number().positive('Nominal payout vendor harus lebih besar dari 0'),
  serviceDescription: z.string().min(3, 'Deskripsi layanan vendor wajib diisi'),
});

export const createTransactionSchema = z.object({
  type: transactionTypeEnum,
  title: z.string().min(3, 'Judul transaksi minimal 3 karakter').max(200),
  description: z.string().max(2000).optional(),
  amount: z.number().positive('Total nominal transaksi harus lebih besar dari 0'),
  categoryId: z.string().uuid('Kategori ID tidak valid'),
  sellerId: z.string().uuid('Seller ID tidak valid'),
  
  // Optional goods details
  shippingAddress: z.string().optional(),
  
  // Optional services milestones
  milestones: z.array(milestoneSchema).optional(),
  
  // Optional events vendor payouts
  vendorPayouts: z.array(eventVendorPayoutSchema).optional(),
});

export const releaseFundsSchema = z.object({
  transactionId: z.string().uuid('ID Transaksi tidak valid'),
  milestoneId: z.string().uuid().optional(),
});

export const createDisputeSchema = z.object({
  transactionId: z.string().uuid('ID Transaksi tidak valid'),
  reason: z.string().min(10, 'Alasan sengketa minimal 10 karakter').max(1000),
  evidenceUrls: z.array(z.string().url()).optional(),
});

export type CreateTransactionInput = z.infer<typeof createTransactionSchema>;
export type ReleaseFundsInput = z.infer<typeof releaseFundsSchema>;
export type CreateDisputeInput = z.infer<typeof createDisputeSchema>;
