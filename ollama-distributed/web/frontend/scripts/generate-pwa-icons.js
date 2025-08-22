#!/usr/bin/env node

/**
 * PWA Icon Generator for OllamaMax
 * Generates all required PWA icons from a base SVG or creates placeholder icons
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Required icon sizes for PWA
const ICON_SIZES = [
  { size: 16, name: 'icon-16x16.png' },
  { size: 32, name: 'icon-32x32.png' },
  { size: 57, name: 'icon-57x57.png' },
  { size: 60, name: 'icon-60x60.png' },
  { size: 70, name: 'icon-70x70.png' },
  { size: 72, name: 'icon-72x72.png' },
  { size: 76, name: 'icon-76x76.png' },
  { size: 96, name: 'icon-96x96.png' },
  { size: 114, name: 'icon-114x114.png' },
  { size: 120, name: 'icon-120x120.png' },
  { size: 128, name: 'icon-128x128.png' },
  { size: 144, name: 'icon-144x144.png' },
  { size: 150, name: 'icon-150x150.png' },
  { size: 152, name: 'icon-152x152.png' },
  { size: 180, name: 'apple-touch-icon.png' },
  { size: 192, name: 'icon-192x192.png' },
  { size: 310, name: 'icon-310x310.png', width: 310, height: 150 },
  { size: 384, name: 'icon-384x384.png' },
  { size: 512, name: 'icon-512x512.png' },
  { size: 512, name: 'icon-512x512-maskable.png', maskable: true }
];

// Splash screen sizes for iOS
const SPLASH_SIZES = [
  { width: 2048, height: 2732, name: 'apple-splash-2048-2732.jpg' },
  { width: 1668, height: 2388, name: 'apple-splash-1668-2388.jpg' },
  { width: 1536, height: 2048, name: 'apple-splash-1536-2048.jpg' },
  { width: 1125, height: 2436, name: 'apple-splash-1125-2436.jpg' },
  { width: 1242, height: 2688, name: 'apple-splash-1242-2688.jpg' },
  { width: 828, height: 1792, name: 'apple-splash-828-1792.jpg' },
  { width: 1242, height: 2208, name: 'apple-splash-1242-2208.jpg' },
  { width: 750, height: 1334, name: 'apple-splash-750-1334.jpg' }
];

const PUBLIC_DIR = path.join(__dirname, '../public');
const ICONS_DIR = path.join(PUBLIC_DIR, 'icons');
const SPLASH_DIR = path.join(PUBLIC_DIR, 'splash');

// Ensure directories exist
function ensureDirectoryExists(dirPath) {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
    console.log(`✅ Created directory: ${dirPath}`);
  }
}

// Generate SVG icon with OllamaMax branding
function generateSVGIcon(size, isMaskable = false) {
  const padding = isMaskable ? size * 0.1 : 0; // 10% padding for maskable icons
  const iconSize = size - (padding * 2);
  const centerOffset = size / 2;
  
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
    </linearGradient>
    ${isMaskable ? `<circle id="mask" cx="${centerOffset}" cy="${centerOffset}" r="${size/2}" fill="white"/>` : ''}
  </defs>
  
  ${isMaskable ? `<circle cx="${centerOffset}" cy="${centerOffset}" r="${size/2}" fill="url(#gradient)"/>` : ''}
  
  <!-- Background circle -->
  <circle cx="${centerOffset}" cy="${centerOffset}" r="${iconSize/2}" fill="${isMaskable ? 'rgba(255,255,255,0.9)' : 'url(#gradient)'}" stroke="${isMaskable ? 'none' : '#ffffff'}" stroke-width="${isMaskable ? 0 : 2}"/>
  
  <!-- OllamaMax "O" Letter -->
  <circle cx="${centerOffset}" cy="${centerOffset}" r="${iconSize * 0.25}" fill="none" stroke="${isMaskable ? '#667eea' : '#ffffff'}" stroke-width="${Math.max(2, iconSize * 0.06)}" opacity="0.9"/>
  
  <!-- Inner dot for "O" -->
  <circle cx="${centerOffset}" cy="${centerOffset}" r="${iconSize * 0.08}" fill="${isMaskable ? '#667eea' : '#ffffff'}" opacity="0.8"/>
  
  <!-- Tech accent marks -->
  <rect x="${centerOffset - iconSize * 0.35}" y="${centerOffset - iconSize * 0.02}" width="${iconSize * 0.15}" height="${iconSize * 0.04}" rx="${iconSize * 0.02}" fill="${isMaskable ? '#764ba2' : '#ffffff'}" opacity="0.6"/>
  <rect x="${centerOffset + iconSize * 0.2}" y="${centerOffset - iconSize * 0.02}" width="${iconSize * 0.15}" height="${iconSize * 0.04}" rx="${iconSize * 0.02}" fill="${isMaskable ? '#764ba2' : '#ffffff'}" opacity="0.6"/>
  
  ${size >= 192 ? `
  <!-- Additional detail for larger icons -->
  <circle cx="${centerOffset - iconSize * 0.2}" cy="${centerOffset - iconSize * 0.2}" r="${iconSize * 0.03}" fill="${isMaskable ? '#667eea' : '#ffffff'}" opacity="0.4"/>
  <circle cx="${centerOffset + iconSize * 0.2}" cy="${centerOffset + iconSize * 0.2}" r="${iconSize * 0.03}" fill="${isMaskable ? '#667eea' : '#ffffff'}" opacity="0.4"/>
  ` : ''}
</svg>`;
}

// Generate placeholder PNG using Canvas (if available) or create SVG
function generateIcon(size, filename, isMaskable = false) {
  const svg = generateSVGIcon(size, isMaskable);
  const svgPath = path.join(ICONS_DIR, filename.replace('.png', '.svg'));
  
  // Write SVG file
  fs.writeFileSync(svgPath, svg);
  
  console.log(`✅ Generated ${filename} (${size}x${size}${isMaskable ? ' maskable' : ''})`);
  
  // Also create a simple HTML file that shows the SVG as PNG for browsers
  const htmlContent = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>${filename}</title></head>
<body style="margin:0;padding:0;display:flex;justify-content:center;align-items:center;min-height:100vh;background:#f0f0f0">
<div style="text-align:center">
<img src="${filename.replace('.png', '.svg')}" width="${size}" height="${size}" alt="OllamaMax Icon"/>
<p>Icon: ${filename} (${size}x${size})</p>
<p>Right-click the icon above and "Save Image As" to get PNG version</p>
</div></body></html>`;
  
  const htmlPath = path.join(ICONS_DIR, filename.replace('.png', '.html'));
  fs.writeFileSync(htmlPath, htmlContent);
}

// Generate splash screen
function generateSplashScreen(width, height, filename) {
  const svg = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="bg-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#667eea;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#764ba2;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background -->
  <rect width="${width}" height="${height}" fill="url(#bg-gradient)"/>
  
  <!-- Center logo -->
  <g transform="translate(${width/2}, ${height/2})">
    <!-- Main circle -->
    <circle r="80" fill="rgba(255,255,255,0.9)" stroke="#ffffff" stroke-width="4"/>
    
    <!-- OllamaMax "O" -->
    <circle r="40" fill="none" stroke="#667eea" stroke-width="8"/>
    <circle r="12" fill="#667eea"/>
    
    <!-- App name -->
    <text x="0" y="140" text-anchor="middle" fill="#ffffff" font-family="system-ui, -apple-system, sans-serif" font-size="24" font-weight="600">OllamaMax</text>
    <text x="0" y="170" text-anchor="middle" fill="rgba(255,255,255,0.8)" font-family="system-ui, -apple-system, sans-serif" font-size="16">Distributed AI Platform</text>
  </g>
</svg>`;
  
  const svgPath = path.join(SPLASH_DIR, filename.replace('.jpg', '.svg'));
  fs.writeFileSync(svgPath, svg);
  
  console.log(`✅ Generated splash screen ${filename} (${width}x${height})`);
}

// Generate favicon
function generateFavicon() {
  const faviconSVG = generateSVGIcon(32);
  const faviconPath = path.join(PUBLIC_DIR, 'favicon.svg');
  fs.writeFileSync(faviconPath, faviconSVG);
  
  // Also create a simple ICO placeholder
  const icoContent = generateSVGIcon(16);
  const icoPath = path.join(PUBLIC_DIR, 'favicon.ico.svg');
  fs.writeFileSync(icoPath, icoContent);
  
  console.log('✅ Generated favicon');
}

// Main execution
function main() {
  console.log('🚀 Generating PWA icons for OllamaMax...\n');
  
  // Ensure directories exist
  ensureDirectoryExists(ICONS_DIR);
  ensureDirectoryExists(SPLASH_DIR);
  
  console.log('📱 Generating app icons...');
  // Generate all app icons
  ICON_SIZES.forEach(({ size, name, maskable }) => {
    generateIcon(size, name, maskable);
  });
  
  console.log('\n🖼️ Generating splash screens...');
  // Generate splash screens
  SPLASH_SIZES.forEach(({ width, height, name }) => {
    generateSplashScreen(width, height, name);
  });
  
  console.log('\n🎯 Generating favicon...');
  // Generate favicon
  generateFavicon();
  
  // Create additional files
  console.log('\n📋 Creating additional PWA files...');
  
  // Create browserconfig.xml for Windows
  const browserConfig = `<?xml version="1.0" encoding="utf-8"?>
<browserconfig>
    <msapplication>
        <tile>
            <square70x70logo src="/icons/icon-70x70.png"/>
            <square150x150logo src="/icons/icon-150x150.png"/>
            <square310x310logo src="/icons/icon-310x310.png"/>
            <TileColor>#2563eb</TileColor>
        </tile>
    </msapplication>
</browserconfig>`;
  
  fs.writeFileSync(path.join(PUBLIC_DIR, 'browserconfig.xml'), browserConfig);
  
  // Create robots.txt
  const robotsTxt = `User-agent: *
Allow: /

Sitemap: https://ollamamax.com/sitemap.xml`;
  
  fs.writeFileSync(path.join(PUBLIC_DIR, 'robots.txt'), robotsTxt);
  
  console.log('✅ Generated browserconfig.xml');
  console.log('✅ Generated robots.txt');
  
  console.log('\n🎉 PWA icon generation completed!');
  console.log('\nGenerated files:');
  console.log(`- ${ICON_SIZES.length} app icons in /public/icons/`);
  console.log(`- ${SPLASH_SIZES.length} splash screens in /public/splash/`);
  console.log('- favicon files in /public/');
  console.log('- browserconfig.xml and robots.txt');
  console.log('\n📝 Note: SVG files are generated as placeholders.');
  console.log('For production, convert SVGs to PNG using a tool like sharp or imagemagick.');
  console.log('\nTo convert SVGs to PNGs:');
  console.log('npm install -g sharp-cli');
  console.log('sharp-cli -f png -q 100 public/icons/*.svg');
}

// Run the generator
if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}

export { generateIcon, generateSplashScreen, generateFavicon };