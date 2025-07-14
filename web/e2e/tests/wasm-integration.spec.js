const { test, expect } = require('@playwright/test');

test.describe('WASM Integration Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForFunction(() => {
      return window.wasmInstance !== undefined;
    }, { timeout: 30000 });
  });

  test('should detect symbols correctly', async ({ page }) => {
    // Test by calling WASM function directly
    const result = await page.evaluate(async () => {
      // Create a simple test image (base64 encoded 1x1 white pixel PNG)
      const testImageBase64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==';
      
      // Call the WASM function
      return window.processGrimoireImage(testImageBase64);
    });

    expect(result).toBeDefined();
    expect(result.success).toBeDefined();
    expect(result.code).toBeDefined();
  });

  test('should handle errors gracefully', async ({ page }) => {
    const result = await page.evaluate(async () => {
      // Invalid base64 data
      return window.processGrimoireImage('invalid-base64');
    });

    expect(result).toBeDefined();
    expect(result.success).toBe(false);
    expect(result.error).toBeDefined();
  });

  test('should return debug information', async ({ page }) => {
    // Click on a sample image
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    // Wait for result
    await page.waitForSelector('.result-section', { state: 'visible' });
    
    // Get the result from the page
    const hasDebugInfo = await page.evaluate(() => {
      const astContent = document.querySelector('#ast-content').textContent;
      return astContent.includes('デバッグ情報');
    });

    expect(hasDebugInfo).toBe(true);
  });

  test('should process all sample images without error', async ({ page }) => {
    const samples = ['hello-world', 'calculator', 'fibonacci', 'loop'];
    
    for (const sample of samples) {
      // Click on sample
      await page.click(`[data-sample="${sample}"]`);
      
      // Execute
      await page.click('#execute-btn');
      
      // Wait for result
      await page.waitForSelector('.result-section', { state: 'visible' });
      
      // Check no error is shown
      const errorSectionVisible = await page.isVisible('.error-section');
      expect(errorSectionVisible).toBe(false);
      
      // Check code was generated
      const codeContent = await page.textContent('#code-content');
      expect(codeContent).toContain('#!/usr/bin/env python3');
      
      // Clear for next test
      await page.reload();
      await page.waitForFunction(() => {
        return window.wasmInstance !== undefined;
      });
    }
  });

  test('should handle star symbols correctly', async ({ page }) => {
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    await page.waitForSelector('.result-section', { state: 'visible' });
    
    // Check if the code contains print statement
    const codeContent = await page.textContent('#code-content');
    expect(codeContent).toMatch(/print\s*\(/);
  });

  test('should handle double circle symbols correctly', async ({ page }) => {
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    await page.waitForSelector('.result-section', { state: 'visible' });
    
    // Check if main entry is created
    const codeContent = await page.textContent('#code-content');
    expect(codeContent).toMatch(/if\s+__name__\s*==\s*["']__main__["']\s*:/);
  });
});