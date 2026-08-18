'use client';

import React, { useEffect, useState } from 'react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { apiService } from '@/services/api.service';

export function CatalogModule() {
  const [categories, setCategories] = useState<Record<string, unknown>[]>([]);
  const [vendors, setVendors] = useState<Record<string, unknown>[]>([]);

  useEffect(() => {
    apiService.getCategories().then((res: unknown) => {
      const data = (res as { data?: Record<string, unknown>[]; result?: Record<string, unknown>[] })?.data || (res as Record<string, unknown>[]) || [];
      setCategories(data);
    }).catch(() => {});
    apiService.listMarketplaceVendors().then((res: unknown) => {
      const data = (res as { data?: Record<string, unknown>[]; result?: Record<string, unknown>[] })?.data || (res as Record<string, unknown>[]) || [];
      setVendors(data);
    }).catch(() => {});
  }, []);

  return (
    <div className="space-y-6">
      <Card className="p-6">
        <CardHeader className="px-0 pt-0">
          <CardTitle className="text-xl">Katalog Kategori (3-Tier Taxonomies)</CardTitle>
          <CardDescription className="text-xs mt-1">
            Taksonomi lengkap untuk Goods, Services, dan Events
          </CardDescription>
        </CardHeader>
        <CardContent className="px-0 pb-0">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {categories.length === 0 ? (
              <p className="text-xs text-zinc-500">Memuat kategori atau menggunakan data mock standar...</p>
            ) : (
              categories.map((cat, idx) => {
                const c = cat as { name?: string; title?: string };
                return (
                  <div key={idx} className="p-4 rounded-xl border border-zinc-200 dark:border-zinc-800 bg-zinc-50 dark:bg-zinc-800/50">
                    <p className="font-bold text-sm text-zinc-900 dark:text-zinc-100">{c.name || c.title || 'Kategori'}</p>
                  </div>
                );
              })
            )}
          </div>
        </CardContent>
      </Card>

      <Card className="p-6">
        <CardHeader className="px-0 pt-0">
          <CardTitle className="text-xl">Vendor Marketplace (Event Organizers)</CardTitle>
          <CardDescription className="text-xs mt-1">
            Daftar vendor katering, sound system, dekorasi, dan venue terverifikasi
          </CardDescription>
        </CardHeader>
        <CardContent className="px-0 pb-0">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {vendors.length === 0 ? (
              <div className="p-4 rounded-xl border border-zinc-200 dark:border-zinc-800">
                <p className="font-bold text-sm">Grand Wedding Catering & Sound</p>
                <Badge variant="success" className="mt-2 text-[10px]">Verified Vendor</Badge>
              </div>
            ) : (
              vendors.map((v, idx) => {
                const vendor = v as { business_name?: string; name?: string };
                return (
                  <div key={idx} className="p-4 rounded-xl border border-zinc-200 dark:border-zinc-800">
                    <p className="font-bold text-sm">{vendor.business_name || vendor.name || 'Vendor'}</p>
                  </div>
                );
              })
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
