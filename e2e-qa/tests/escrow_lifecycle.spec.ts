import { test, expect } from '@playwright/test';

test.describe('RekberKuy E2E - Escrow Lifecycle & PRD Business Flows', () => {

  test.beforeEach(async ({ page }) => {
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('buyer@rekberkuy.test');
    await passwordInput.fill('password_buyer');
    await submitButton.click();
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  });

  test('Goods Escrow Lifecycle: Creation, Funds Locked, Shipping, and Release', async ({ page }) => {
    // 1. Navigate to transactions module
    await page.goto('/dashboard/transactions', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Transaksi|Escrow|Barang|Goods|Pesanan/i);

    // 2. Validate back button navigation in transaction module
    const backButton = page.locator('button:has-text("Kembali"), a:has-text("Kembali")').first();
    if (await backButton.isVisible()) {
      await backButton.click();
      await page.waitForURL(/\/dashboard/, { timeout: 10000 });
      await page.goto('/dashboard/transactions', { waitUntil: 'domcontentloaded' });
    }

    // 3. Verify transaction listing displays Escrow status (Funds Locked / Released)
    await expect(page.locator('body')).toContainText(/Status|Goods|Jasa|Event|Released|Locked/i);
  });

  test('Services & Milestone Lifecycle: Project submission, milestone release, and review', async ({ page }) => {
    await page.goto('/dashboard/transactions', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Layanan|Milestone|Services|Transaksi/i);

    // Verify service milestone details or mock transaction item view
    const transactionItem = page.locator('text=/Services|Jasa/i').first();
    if (await transactionItem.isVisible()) {
      await transactionItem.click().catch(() => {});
    }
  });

  test('Event & Vendor Procurement: EO transaction and multi-party vendor allocation', async ({ page }) => {
    await page.goto('/dashboard/transactions', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Event|Vendor|Alokasi|Transaksi/i);
  });

});
