#!/usr/bin/env node

/**
 * Build script for Social Network Messenger
 * Handles cross-platform builds and distribution
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
const platform = args[0] || 'all';

console.log('🚀 Building Social Network Messenger...');

// Check if we're in the right directory
if (!fs.existsSync('package.json')) {
    console.error('❌ Error: package.json not found. Please run this script from the desktop-messenger directory.');
    process.exit(1);
}

// Check if node_modules exists
if (!fs.existsSync('node_modules')) {
    console.log('📦 Installing dependencies...');
    try {
        execSync('npm install', { stdio: 'inherit' });
    } catch (error) {
        console.error('❌ Error installing dependencies:', error.message);
        process.exit(1);
    }
}

// Build functions
const builds = {
    windows: () => {
        console.log('🪟 Building for Windows...');
        execSync('npm run build:win', { stdio: 'inherit' });
    },
    
    mac: () => {
        console.log('🍎 Building for macOS...');
        execSync('npm run build:mac', { stdio: 'inherit' });
    },
    
    linux: () => {
        console.log('🐧 Building for Linux...');
        execSync('npm run build:linux', { stdio: 'inherit' });
    },
    
    all: () => {
        console.log('🌍 Building for all platforms...');
        execSync('npm run build', { stdio: 'inherit' });
    }
};

// Validate platform
if (!builds[platform]) {
    console.error(`❌ Error: Unknown platform "${platform}". Available options: windows, mac, linux, all`);
    process.exit(1);
}

try {
    // Run the build
    builds[platform]();
    
    console.log('✅ Build completed successfully!');
    console.log('📁 Output files are in the dist/ directory');
    
    // List output files
    const distDir = path.join(__dirname, 'dist');
    if (fs.existsSync(distDir)) {
        console.log('\n📋 Generated files:');
        const files = fs.readdirSync(distDir);
        files.forEach(file => {
            const filePath = path.join(distDir, file);
            const stats = fs.statSync(filePath);
            const size = (stats.size / 1024 / 1024).toFixed(2);
            console.log(`   ${file} (${size} MB)`);
        });
    }
    
} catch (error) {
    console.error('❌ Build failed:', error.message);
    process.exit(1);
}

console.log('\n🎉 Build process completed!');
console.log('💡 Tip: You can now distribute the files in the dist/ directory');
