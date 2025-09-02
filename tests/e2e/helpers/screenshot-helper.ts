import { Page, Locator } from '@playwright/test';
import fs from 'fs/promises';
import path from 'path';

/**
 * Screenshot Helper for OllamaMax Platform Testing
 * 
 * Provides utilities for:
 * - Full page screenshots
 * - Element screenshots
 * - Responsive screenshots across devices
 * - Visual regression testing support
 * - Screenshot comparison and analysis
 */

export interface ScreenshotOptions {
  fullPage?: boolean;
  clip?: { x: number; y: number; width: number; height: number };
  quality?: number;
  type?: 'png' | 'jpeg';
  animations?: 'disabled' | 'allow';
  caret?: 'hide' | 'initial';
  scale?: 'css' | 'device';
  threshold?: number;
  thresholdType?: 'percent' | 'pixels';
}

export class ScreenshotHelper {
  private screenshotDir: string;

  constructor(private page: Page) {
    this.screenshotDir = path.join('reports', 'screenshots');
  }

  /**
   * Capture full page screenshot
   */
  async captureFullPage(
    name: string, 
    options: ScreenshotOptions = {}
  ): Promise<string> {
    await this.ensureScreenshotDir();
    
    const timestamp = this.getTimestamp();
    const filename = `${name}-${timestamp}.png`;
    const filepath = path.join(this.screenshotDir, filename);
    
    const defaultOptions = {
      fullPage: true,
      animations: 'disabled' as const,
      caret: 'hide' as const,
      ...options
    };
    
    await this.page.screenshot({
      path: filepath,
      ...defaultOptions
    });
    
    console.log(`📸 Screenshot saved: ${filepath}`);
    return filepath;
  }

  /**
   * Capture element screenshot
   */
  async captureElement(
    element: Locator, 
    name: string, 
    options: ScreenshotOptions = {}
  ): Promise<string> {
    await this.ensureScreenshotDir();
    
    const timestamp = this.getTimestamp();
    const filename = `element-${name}-${timestamp}.png`;
    const filepath = path.join(this.screenshotDir, filename);
    
    const defaultOptions = {
      animations: 'disabled' as const,
      ...options
    };
    
    await element.screenshot({
      path: filepath,
      ...defaultOptions
    });
    
    console.log(`📸 Element screenshot saved: ${filepath}`);
    return filepath;
  }

  /**
   * Capture responsive screenshots across multiple viewports
   */
  async captureResponsive(name: string): Promise<string[]> {
    const viewports = [
      { width: 375, height: 812, name: 'mobile-portrait' },
      { width: 667, height: 375, name: 'mobile-landscape' },
      { width: 768, height: 1024, name: 'tablet-portrait' },
      { width: 1024, height: 768, name: 'tablet-landscape' },
      { width: 1280, height: 720, name: 'desktop-small' },
      { width: 1920, height: 1080, name: 'desktop-large' }
    ];
    
    const screenshots = [];
    const originalViewport = this.page.viewportSize();
    
    for (const viewport of viewports) {
      await this.page.setViewportSize({ 
        width: viewport.width, 
        height: viewport.height 
      });
      
      // Wait for layout to settle
      await this.page.waitForTimeout(1000);
      
      const filename = await this.captureFullPage(`${name}-${viewport.name}`);
      screenshots.push(filename);
    }
    
    // Restore original viewport
    if (originalViewport) {
      await this.page.setViewportSize(originalViewport);
    }
    
    console.log(`📸 Responsive screenshots captured for: ${name}`);
    return screenshots;
  }

  /**
   * Capture screenshot with annotations
   */
  async captureWithAnnotations(
    name: string,
    annotations: Array<{
      x: number;
      y: number;
      width?: number;
      height?: number;
      text?: string;
      color?: string;
    }>
  ): Promise<string> {
    // Add visual annotations to the page
    await this.page.evaluate((annotations) => {
      const style = document.createElement('style');
      style.innerHTML = `
        .test-annotation {
          position: absolute;
          border: 2px solid #ff0000;
          background: rgba(255, 0, 0, 0.1);
          color: #ff0000;
          font-family: Arial, sans-serif;
          font-size: 12px;
          font-weight: bold;
          padding: 2px 4px;
          z-index: 10000;
          pointer-events: none;
        }
      `;
      document.head.appendChild(style);
      
      annotations.forEach((annotation, index) => {
        const div = document.createElement('div');
        div.className = 'test-annotation';
        div.style.left = `${annotation.x}px`;
        div.style.top = `${annotation.y}px`;
        
        if (annotation.width) div.style.width = `${annotation.width}px`;
        if (annotation.height) div.style.height = `${annotation.height}px`;
        if (annotation.color) {
          div.style.borderColor = annotation.color;
          div.style.color = annotation.color;
        }
        
        div.textContent = annotation.text || `Annotation ${index + 1}`;
        document.body.appendChild(div);
      });
    }, annotations);
    
    const filepath = await this.captureFullPage(`${name}-annotated`);
    
    // Clean up annotations
    await this.page.evaluate(() => {
      document.querySelectorAll('.test-annotation').forEach(el => el.remove());
      document.querySelectorAll('style').forEach(el => {
        if (el.innerHTML.includes('test-annotation')) {
          el.remove();
        }
      });
    });
    
    return filepath;
  }

  /**
   * Capture screenshot comparison (before/after)
   */
  async captureComparison(
    name: string,
    beforeAction: () => Promise<void>,
    afterAction: () => Promise<void>
  ): Promise<{ before: string; after: string }> {
    // Capture before screenshot
    const beforePath = await this.captureFullPage(`${name}-before`);
    
    // Execute the action
    await beforeAction();
    
    // Wait for changes to settle
    await this.page.waitForTimeout(1000);
    
    // Capture after screenshot
    const afterPath = await this.captureFullPage(`${name}-after`);
    
    await afterAction();
    
    console.log(`📸 Comparison screenshots captured for: ${name}`);
    return { before: beforePath, after: afterPath };
  }

  /**
   * Capture error state screenshot
   */
  async captureError(
    name: string,
    error: Error,
    additionalInfo?: any
  ): Promise<string> {
    await this.ensureScreenshotDir();
    
    const timestamp = this.getTimestamp();
    const filename = `error-${name}-${timestamp}.png`;
    const filepath = path.join(this.screenshotDir, filename);
    
    // Add error overlay to the page
    await this.page.evaluate(({ errorMessage, additionalInfo }) => {
      const overlay = document.createElement('div');
      overlay.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(255, 0, 0, 0.1);
        z-index: 9999;
        pointer-events: none;
      `;
      
      const errorBox = document.createElement('div');
      errorBox.style.cssText = `
        position: fixed;
        top: 20px;
        left: 20px;
        right: 20px;
        background: #ff4444;
        color: white;
        padding: 20px;
        border-radius: 8px;
        font-family: monospace;
        font-size: 14px;
        z-index: 10000;
        max-height: 300px;
        overflow-y: auto;
      `;
      
      errorBox.innerHTML = `
        <h3>Test Error</h3>
        <p><strong>Message:</strong> ${errorMessage}</p>
        <p><strong>Time:</strong> ${new Date().toISOString()}</p>
        ${additionalInfo ? `<p><strong>Additional Info:</strong> ${JSON.stringify(additionalInfo, null, 2)}</p>` : ''}
      `;
      
      overlay.appendChild(errorBox);
      document.body.appendChild(overlay);
    }, {
      errorMessage: error.message,
      additionalInfo
    });
    
    await this.page.screenshot({
      path: filepath,
      fullPage: true
    });
    
    // Clean up error overlay
    await this.page.evaluate(() => {
      document.querySelectorAll('[style*="z-index: 9999"]').forEach(el => el.remove());
    });
    
    console.log(`📸 Error screenshot saved: ${filepath}`);
    return filepath;
  }

  /**
   * Capture network activity screenshot
   */
  async captureWithNetworkInfo(name: string): Promise<string> {
    // Get network information
    const networkInfo = await this.page.evaluate(() => {
      const resources = performance.getEntriesByType('resource');
      const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
      
      return {
        resourceCount: resources.length,
        loadTime: navigation ? Math.round(navigation.loadEventEnd - navigation.fetchStart) : 0,
        domReady: navigation ? Math.round(navigation.domContentLoadedEventEnd - navigation.fetchStart) : 0,
        totalSize: resources.reduce((sum: number, resource: any) => sum + (resource.transferSize || 0), 0)
      };
    });
    
    // Add network info overlay
    await this.page.evaluate((info) => {
      const overlay = document.createElement('div');
      overlay.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        background: rgba(0, 0, 0, 0.8);
        color: white;
        padding: 15px;
        border-radius: 8px;
        font-family: monospace;
        font-size: 12px;
        z-index: 10000;
        pointer-events: none;
      `;
      
      overlay.innerHTML = `
        <h4>Network Info</h4>
        <p>Resources: ${info.resourceCount}</p>
        <p>Load Time: ${info.loadTime}ms</p>
        <p>DOM Ready: ${info.domReady}ms</p>
        <p>Total Size: ${Math.round(info.totalSize / 1024)}KB</p>
        <p>Time: ${new Date().toLocaleTimeString()}</p>
      `;
      
      document.body.appendChild(overlay);
    }, networkInfo);
    
    const filepath = await this.captureFullPage(`${name}-network`);
    
    // Clean up overlay
    await this.page.evaluate(() => {
      document.querySelectorAll('[style*="z-index: 10000"]').forEach(el => el.remove());
    });
    
    return filepath;
  }

  /**
   * Create screenshot gallery HTML
   */
  async createGallery(
    screenshots: string[], 
    title: string = 'Test Screenshots'
  ): Promise<string> {
    const galleryHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>${title}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .gallery { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .screenshot { background: white; border-radius: 8px; padding: 15px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .screenshot img { width: 100%; height: auto; border-radius: 4px; }
        .screenshot h3 { margin: 0 0 10px 0; color: #333; }
        .screenshot p { margin: 5px 0; color: #666; font-size: 14px; }
        .header { text-align: center; margin-bottom: 30px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>${title}</h1>
        <p>Generated on ${new Date().toLocaleString()}</p>
        <p>Total screenshots: ${screenshots.length}</p>
    </div>
    <div class="gallery">
        ${screenshots.map((screenshot, index) => {
          const filename = path.basename(screenshot);
          const relativePath = path.relative(path.dirname(screenshot), screenshot);
          
          return `
            <div class="screenshot">
                <h3>${filename}</h3>
                <img src="${relativePath}" alt="${filename}" />
                <p>File: ${filename}</p>
                <p>Index: ${index + 1}</p>
            </div>
          `;
        }).join('')}
    </div>
</body>
</html>
    `;
    
    const galleryPath = path.join(this.screenshotDir, 'gallery.html');
    await fs.writeFile(galleryPath, galleryHTML);
    
    console.log(`📸 Screenshot gallery created: ${galleryPath}`);
    return galleryPath;
  }

  /**
   * Utility methods
   */
  private async ensureScreenshotDir(): Promise<void> {
    await fs.mkdir(this.screenshotDir, { recursive: true }).catch(() => {});
  }
  
  private getTimestamp(): string {
    return new Date().toISOString()
      .replace(/[:.]/g, '-')
      .replace('T', '_')
      .split('.')[0];
  }
  
  /**
   * Clean up old screenshots
   */
  async cleanupOldScreenshots(daysOld: number = 7): Promise<number> {
    try {
      const files = await fs.readdir(this.screenshotDir);
      const cutoffTime = Date.now() - (daysOld * 24 * 60 * 60 * 1000);
      let deletedCount = 0;
      
      for (const file of files) {
        if (file.endsWith('.png') || file.endsWith('.jpg')) {
          const filepath = path.join(this.screenshotDir, file);
          const stats = await fs.stat(filepath);
          
          if (stats.mtime.getTime() < cutoffTime) {
            await fs.unlink(filepath);
            deletedCount++;
          }
        }
      }
      
      console.log(`🧹 Cleaned up ${deletedCount} old screenshots`);
      return deletedCount;
    } catch (error) {
      console.warn('Failed to cleanup old screenshots:', error);
      return 0;
    }
  }
}