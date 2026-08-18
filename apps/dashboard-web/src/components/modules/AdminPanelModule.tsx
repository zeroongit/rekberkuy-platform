'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { apiService } from '@/services/api.service';
import { notify } from '@/components/providers/ToastProvider';
import { ShieldCheck, CheckCircle, XCircle } from 'lucide-react';

interface KycItem {
  id: string;
  full_name: string;
  nik: string;
  ai_score?: number;
}

export function AdminPanelModule() {
  const [pendingKycs, setPendingKycs] = useState<KycItem[]>([]);

  useEffect(() => {
    apiService
      .adminGetPendingKYCs()
      .then((res: unknown) => {
        const data = (res as { data?: KycItem[]; result?: KycItem[] })?.data || (res as KycItem[]) || [];
        setPendingKycs(data);
      })
      .catch(() => {});
  }, []);

  const handleReviewKyc = async (id: string, approved: boolean) => {
    try {
      await apiService.adminReviewKYC(id, { approved, reason: approved ? 'Dokumen valid dan sesuai KTP' : 'Dokumen kurang jelas' });
      notify.success(approved ? 'KYC berhasil disetujui!' : 'KYC ditolak.');
      setPendingKycs((prev) => prev.filter((k) => k.id !== id));
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal mereview KYC';
      notify.error(msg);
    }
  };

  return (
    <Card className="p-6">
      <CardHeader className="px-0 pt-0">
        <div className="flex items-center space-x-3">
          <div className="p-2 rounded-xl bg-purple-50 dark:bg-purple-950/50 text-purple-600 dark:text-purple-400">
            <ShieldCheck className="h-6 w-6" />
          </div>
          <div>
            <CardTitle className="text-xl">Panel Admin & Mediasi KYC</CardTitle>
            <CardDescription className="text-xs mt-0.5">
              Tinjauan verifikasi KYC (AI scoring sebagai referensi, keputusan akhir oleh Admin)
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="px-0 pb-0 space-y-4">
        <div className="space-y-3">
          <h4 className="text-sm font-bold">Daftar Pengajuan KYC Tertunda</h4>
          {pendingKycs.length === 0 ? (
            <p className="text-xs text-zinc-500">Tidak ada pengajuan KYC tertunda saat ini.</p>
          ) : (
            pendingKycs.map((kyc) => (
              <div key={kyc.id} className="p-4 rounded-xl border border-zinc-200 dark:border-zinc-800 flex items-center justify-between">
                <div>
                  <p className="font-bold text-sm">{kyc.full_name}</p>
                  <p className="text-xs text-zinc-500 font-mono">NIK: {kyc.nik}</p>
                  {kyc.ai_score !== undefined && (
                    <Badge variant="secondary" className="mt-1 text-[10px]">
                      AI Score: {kyc.ai_score}
                    </Badge>
                  )}
                </div>
                <div className="flex space-x-2">
                  <Button size="sm" variant="default" onClick={() => handleReviewKyc(kyc.id, true)} className="text-xs">
                    <CheckCircle className="h-3.5 w-3.5 mr-1" />
                    Setujui
                  </Button>
                  <Button size="sm" variant="destructive" onClick={() => handleReviewKyc(kyc.id, false)} className="text-xs">
                    <XCircle className="h-3.5 w-3.5 mr-1" />
                    Tolak
                  </Button>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
