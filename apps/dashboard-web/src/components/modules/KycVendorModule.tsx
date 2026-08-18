'use client';

import React, { useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { notify } from '@/components/providers/ToastProvider';
import { apiService } from '@/services/api.service';

export function KycVendorModule() {
  const [nik, setNik] = useState('');
  const [fullName, setFullName] = useState('');
  const [bankName, setBankName] = useState('');
  const [bankAccountNumber, setBankAccountNumber] = useState('');
  const [bankAccountHolder, setBankAccountHolder] = useState('');

  const handleKycSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiService.submitKyc({
        nik,
        full_name: fullName,
        id_card_photo_url: 'https://example.com/ktp.jpg',
        selfie_photo_url: 'https://example.com/selfie.jpg',
        bank_name: bankName,
        bank_account_number: bankAccountNumber,
        bank_account_holder: bankAccountHolder,
      });
      notify.success('Pengajuan KYC berhasil dikirim! Menunggu tinjauan admin.');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal mengirim KYC';
      notify.error(msg);
    }
  };

  return (
    <Card className="p-6">
      <CardHeader className="px-0 pt-0">
        <CardTitle className="text-xl">Verifikasi Identitas (KYC)</CardTitle>
        <CardDescription className="text-xs mt-1">
          Dapatkan status Verified Merchant untuk komisi lebih rendah dan batasan transaksi lebih tinggi
        </CardDescription>
      </CardHeader>
      <CardContent className="px-0 pb-0">
        <form onSubmit={handleKycSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-bold mb-1">NIK KTP (16 Digit)</label>
              <Input
                type="text"
                maxLength={16}
                value={nik}
                onChange={(e) => setNik(e.target.value)}
                placeholder="3271xxxxxxxxxxxx"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-bold mb-1">Nama Sesuai KTP</label>
              <Input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                placeholder="Budi Santoso"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-bold mb-1">Nama Bank</label>
              <Input
                type="text"
                value={bankName}
                onChange={(e) => setBankName(e.target.value)}
                placeholder="BCA / Mandiri / BNI"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-bold mb-1">Nomor Rekening</label>
              <Input
                type="text"
                value={bankAccountNumber}
                onChange={(e) => setBankAccountNumber(e.target.value)}
                placeholder="1234567890"
                required
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-bold mb-1">Nama Pemilik Rekening</label>
            <Input
              type="text"
              value={bankAccountHolder}
              onChange={(e) => setBankAccountHolder(e.target.value)}
              placeholder="Budi Santoso"
              required
            />
          </div>

          <Button type="submit" className="text-xs font-bold">
            Kirim Pengajuan KYC
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
