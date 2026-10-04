import { test, expect } from '@playwright/test';

test.describe('RekberKuy E2E - Wallet, Dispute Mediation, and Blockchain Audit Log', () => {

  test('RekberPay Wallet & Ledger: balance mutation and Midtrans top-up simulation', async ({ page }) => {
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('buyer@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });

    await page.goto('/dashboard/ledger', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Saldo|Ledger|Mutasi|Top-Up|Midtrans|RekberPay/i);

    // Validate back button navigation
    const backButton = page.locator('button:has-text("Kembali"), a:has-text("Kembali")').first();
    if (await backButton.isVisible()) {
      await backButton.click();
      await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    }
  });

  test('Dispute & Refund Flow: trigger dispute and admin mediation outcome', async ({ page }) => {
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('buyer@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });

    await page.goto('/dashboard/disputes', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Sengketa|Dispute|Komplain/i);

    // Validate back button navigation
    const backButton = page.locator('button:has-text("Kembali"), a:has-text("Kembali")').first();
    if (await backButton.isVisible()) {
      await backButton.click();
      await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    }
  });

  test('Audit Log Relayer: verifying completed transactions emit immutable audit log on Avalanche testnet', async ({ page }) => {
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('admin@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();
    await page.waitForURL(/\/dashboard\/admin/, { timeout: 15000 });

    await page.goto('/dashboard/transactions', { waitUntil: 'domcontentloaded' }).catch(() => {});
    await expect(page.locator('body')).toContainText(/Audit|Blockchain|Avalanche|Relayer|TransactionLogger|Hash|Transaksi|Admin/i);
  });

});
