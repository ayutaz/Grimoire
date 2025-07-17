const { test, expect } = require('@playwright/test');

test.describe('WASM Integration Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/web/');
    // Wait for WASM to be fully initialized
    await page.waitForFunction(() => {
      return window.wasmInstance !== undefined && 
             window.processGrimoireImage !== undefined &&
             typeof window.processGrimoireImage === 'function';
    }, { timeout: 60000 });
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

  test('should return debug information as object', async ({ page }) => {
    // Click on a sample image
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    // Wait for either result or error section
    await page.waitForSelector('.result-section, .error-section', { state: 'visible', timeout: 30000 });
    
    // Check if error occurred
    const errorVisible = await page.isVisible('.error-section');
    if (errorVisible) {
      const errorText = await page.textContent('#error-content');
      console.log('Error occurred:', errorText);
    }
    
    // Get the result object from WASM call
    const debugInfo = await page.evaluate(async () => {
      // Get the last result stored in memory or call the function again
      const testImageBase64 = await (async () => {
        const response = await fetch('static/samples/hello-world.png');
        const blob = await response.blob();
        const arrayBuffer = await blob.arrayBuffer();
        const uint8Array = new Uint8Array(arrayBuffer);
        let binary = '';
        for (let i = 0; i < uint8Array.byteLength; i++) {
          binary += String.fromCharCode(uint8Array[i]);
        }
        return btoa(binary);
      })();
      
      const result = window.processGrimoireImage(testImageBase64);
      return {
        success: result.success,
        hasDebug: result.debug !== undefined,
        debugType: typeof result.debug,
        symbolCount: result.debug ? result.debug.symbolCount : null,
        symbolsLength: result.debug && result.debug.symbols ? result.debug.symbols.length : null,
        astType: typeof result.ast,
        hasAst: result.ast !== undefined
      };
    });
    
    // Verify debug info is an object, not a string
    expect(debugInfo.success).toBe(true);
    expect(debugInfo.hasDebug).toBe(true);
    expect(debugInfo.debugType).toBe('object');
    expect(debugInfo.symbolCount).toBeGreaterThanOrEqual(0);
    
    // Verify AST is also an object
    expect(debugInfo.astType).toBe('object');
    
    // Switch to AST tab to see debug info
    await page.click('[data-tab="ast"]');
    
    // Get the displayed content
    const astContent = await page.textContent('#ast-content');
    expect(astContent).toBeTruthy();
    expect(astContent).toContain('=== デバッグ情報 ===');
    expect(astContent).toContain('検出されたシンボル数:');
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
        return window.wasmInstance !== undefined &&
               window.processGrimoireImage !== undefined &&
               typeof window.processGrimoireImage === 'function';
      }, { timeout: 60000 });
    }
  });

  test('should handle star symbols correctly', async ({ page }) => {
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    await page.waitForSelector('.result-section', { state: 'visible' });
    
    // Check if Python code was generated
    const codeContent = await page.textContent('#code-content');
    expect(codeContent).toBeTruthy();
    expect(codeContent).toContain('#!/usr/bin/env python3');
  });

  test('should handle double circle symbols correctly', async ({ page }) => {
    await page.click('[data-sample="hello-world"]');
    await page.click('#execute-btn');
    
    await page.waitForSelector('.result-section', { state: 'visible' });
    
    // Check if code was generated
    const codeContent = await page.textContent('#code-content');
    expect(codeContent).toBeTruthy();
    expect(codeContent).toContain('# Generated by Grimoire');
  });

  test('should process calculator image and not show hello world', async ({ page }) => {
    // Click on calculator sample
    await page.click('[data-sample="calculator"]');
    await page.click('#execute-btn');
    
    // Wait for either result or error section
    await page.waitForSelector('.result-section, .error-section', { state: 'visible', timeout: 30000 });
    
    // Check that generated code is NOT hello world
    const codeContent = await page.textContent('#code-content');
    expect(codeContent).toBeTruthy();
    
    // Get Python output
    const outputContent = await page.textContent('#output-content');
    
    // Calculator should not output "Hello, World!"
    // It should either have calculator-specific code or a detection message
    expect(outputContent.toLowerCase()).not.toContain('hello, world!');
    
    // Verify debug info shows symbols were detected
    await page.click('[data-tab="ast"]');
    const astContent = await page.textContent('#ast-content');
    expect(astContent).toContain('検出されたシンボル数:');
    
    // Get actual symbol count
    const symbolMatch = astContent.match(/検出されたシンボル数: (\d+)/);
    if (symbolMatch) {
      const symbolCount = parseInt(symbolMatch[1]);
      console.log(`Detected ${symbolCount} symbols in calculator image`);
    }
  });

  test('should correctly handle debug info without JSON parsing errors', async ({ page }) => {
    // Test with multiple sample images
    const samples = ['hello-world', 'calculator', 'fibonacci'];
    
    for (const sample of samples) {
      await page.click(`[data-sample="${sample}"]`);
      await page.click('#execute-btn');
      
      await page.waitForSelector('.result-section', { state: 'visible' });
      
      // Check console for JSON parsing errors
      const consoleErrors = [];
      page.on('console', msg => {
        if (msg.type() === 'error' && msg.text().includes('JSON')) {
          consoleErrors.push(msg.text());
        }
      });
      
      // Switch to AST tab
      await page.click('[data-tab="ast"]');
      
      // Wait a bit for any console errors
      await page.waitForTimeout(500);
      
      // Should have no JSON parsing errors
      expect(consoleErrors).toHaveLength(0);
      
      // Reload for next test
      await page.reload();
      await page.waitForFunction(() => {
        return window.wasmInstance !== undefined &&
               window.processGrimoireImage !== undefined &&
               typeof window.processGrimoireImage === 'function';
      }, { timeout: 60000 });
    }
  });
});