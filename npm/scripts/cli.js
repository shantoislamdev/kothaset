#!/usr/bin/env node

/**
 * KothaSet CLI wrapper
 * This wrapper script executes the downloaded KothaSet binary
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const BINARY_NAME = process.platform === 'win32' ? 'kothaset.exe' : 'kothaset';
const BIN_DIR = path.join(__dirname, '..', 'bin');
const BINARY_PATH = path.join(BIN_DIR, BINARY_NAME);

// Check if binary exists
if (!fs.existsSync(BINARY_PATH)) {
  console.error('\n❌ KothaSet binary not found.');
  console.error('   Expected location:', BINARY_PATH);
  console.error('\n   Please try reinstalling the package:');
  console.error('   npm uninstall -g kothaset && npm install -g kothaset\n');
  process.exit(1);
}

// Spawn the binary with all arguments
const child = spawn(BINARY_PATH, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: false
});

child.on('error', (err) => {
  console.error('\n❌ Failed to start KothaSet:', err.message);
  process.exit(1);
});

child.on('exit', (code) => {
  process.exit(code || 0);
});
