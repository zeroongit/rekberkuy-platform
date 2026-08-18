import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(8, 'Password minimal 8 karakter'),
});

export const registerSchema = z.object({
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(8, 'Password minimal 8 karakter'),
  fullName: z.string().min(2, 'Nama lengkap minimal 2 karakter').max(100),
  phone: z.string().min(10, 'Nomor telepon minimal 10 digit').max(15).optional(),
});

export const kycSubmissionSchema = z.object({
  nik: z.string().length(16, 'NIK KTP harus tepat 16 digit'),
  fullName: z.string().min(2, 'Nama lengkap sesuai KTP wajib diisi'),
  idCardPhotoUrl: z.string().url('URL foto KTP tidak valid'),
  selfiePhotoUrl: z.string().url('URL foto selfie dengan KTP tidak valid'),
  bankName: z.string().min(2, 'Nama bank wajib diisi'),
  bankAccountNumber: z.string().min(5, 'Nomor rekening bank wajib diisi'),
  bankAccountHolder: z.string().min(2, 'Nama pemilik rekening wajib diisi'),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type KYCSubmissionInput = z.infer<typeof kycSubmissionSchema>;
