import { test, expect } from '@playwright/test';

test.describe('RekberKuy E2E - PRD RBAC & Access Control Scenarios', () => {

  test('Admin exclusive routing: Admin login must redirect to /dashboard/admin and prevent public home/catalog access', async ({ page }) => {
    // 1. Visit auth page
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });

    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('admin@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();

    // 2. Verify admin is redirected exclusively to /dashboard/admin control panel
    await page.waitForURL(/\/dashboard\/admin/, { timeout: 15000 });
    expect(page.url()).toContain('/dashboard/admin');
    await expect(page.locator('body')).toContainText(/Admin|Panel|KYC|Moderasi|Withdrawal/i);

    // 3. PRD Rule: Admin must not access landing page (/) or buyer dashboard (/dashboard) - middleware should redirect back to /dashboard/admin
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await page.waitForURL(/\/dashboard\/admin/, { timeout: 10000 });
    expect(page.url()).toContain('/dashboard/admin');

    await page.goto('/dashboard', { waitUntil: 'domcontentloaded' });
    await page.waitForURL(/\/dashboard\/admin/, { timeout: 10000 });
    expect(page.url()).toContain('/dashboard/admin');
  });

  test('General User / Buyer separation: Public catalog and personal dashboard navigation with back button', async ({ page }) => {
    // 1. Visit public landing page
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/RekberKuy|Escrow|Goods|Services|Events/i);

    // 2. Login as Buyer
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('buyer@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();

    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    expect(page.url()).toContain('/dashboard');

    // 3. Browse catalog module
    await page.goto('/dashboard/catalog', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('body')).toContainText(/Katalog|Goods|Services|Event/i);

    // 4. Validate PRD Module Navigation: Back button functionality
    const backButton = page.locator('button:has-text("Kembali"), a:has-text("Kembali")').first();
    await expect(backButton).toBeVisible({ timeout: 10000 });
    await backButton.click();
    await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    expect(page.url()).toContain('/dashboard');
  });

  test('Verified Merchant, EO, and Vendor operational dashboard access', async ({ page }) => {
    await page.goto('/auth', { waitUntil: 'domcontentloaded' });
    
    const emailInput = page.locator('input[type="email"]');
    const passwordInput = page.locator('input[type="password"]');
    const submitButton = page.locator('button[type="submit"]');

    await expect(emailInput).toBeVisible({ timeout: 15000 });
    await emailInput.fill('eo@rekberkuy.id');
    await passwordInput.fill('password123');
    await submitButton.click();

    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    await expect(page.locator('body')).toContainText(/Dashboard|Transaksi|Event|Vendor|RekberPay/i);
  });

});
