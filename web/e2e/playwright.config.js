const { defineConfig, devices } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests',
  timeout: process.env.CI ? 60 * 1000 : 30 * 1000,  // 60s timeout in CI
  expect: {
    timeout: process.env.CI ? 20000 : 10000  // 20s expect timeout in CI
  },
  fullyParallel: false,  // Run tests sequentially in CI to avoid resource contention
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 0 : 0,  // No retries to avoid hanging
  workers: 1,  // Always use single worker
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'html',  // Enhanced CI reporting
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: process.env.CI ? 'retain-on-failure' : 'off',  // Record video in CI for debugging
    actionTimeout: process.env.CI ? 30000 : 10000,  // 30s action timeout in CI
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: {
    command: 'npx http-server .. -p 8080 -c-1 --no-dotfiles',
    port: 8080,
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,  // 120 seconds timeout for server startup in CI
    stdout: 'pipe',
    stderr: 'pipe',
  },
});