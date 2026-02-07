#!/usr/bin/env node

/**
 * KothaSet postinstall script
 * Downloads the correct binary for the current platform from GitHub Releases
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const PACKAGE = require('../package.json');
const VERSION = PACKAGE.version;
const REPO = 'shantoislamdev/kothaset';
const BIN_DIR = path.join(__dirname, '..', 'bin');
const BINARY_NAME = process.platform === 'win32' ? 'kothaset.exe' : 'kothaset';

// Platform/arch mapping
const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64',
};

function getPlatformArch() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    throw new Error(
      `Unsupported platform: ${process.platform} ${process.arch}\n` +
      `Supported: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64`
    );
  }

  return { platform, arch };
}

function getDownloadUrl() {
  const { platform, arch } = getPlatformArch();
  const ext = platform === 'windows' ? 'zip' : 'tar.gz';
  const filename = `kothaset_${VERSION}_${platform}_${arch}.${ext}`;
  return `https://github.com/${REPO}/releases/download/v${VERSION}/${filename}`;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    console.log(`Downloading from: ${url}`);
    
    const file = fs.createWriteStream(dest);
    
    const request = https.get(url, (response) => {
      // Handle redirects
      if (response.statusCode === 302 || response.statusCode === 301) {
        file.close();
        fs.unlinkSync(dest);
        download(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }

      if (response.statusCode !== 200) {
        file.close();
        fs.unlinkSync(dest);
        reject(new Error(`Failed to download: HTTP ${response.statusCode}`));
        return;
      }

      const total = parseInt(response.headers['content-length'], 10);
      let downloaded = 0;

      response.on('data', (chunk) => {
        downloaded += chunk.length;
        if (total) {
          const percent = Math.round((downloaded / total) * 100);
          process.stdout.write(`\rDownloading... ${percent}%`);
        }
      });

      response.pipe(file);

      file.on('finish', () => {
        file.close();
        console.log('\nDownload complete.');
        resolve();
      });
    });

    request.on('error', (err) => {
      file.close();
      fs.unlinkSync(dest);
      reject(err);
    });
  });
}

function extract(archive, dest) {
  const { platform } = getPlatformArch();
  
  console.log('Extracting...');
  
  if (platform === 'windows') {
    // Use PowerShell to extract zip on Windows
    execSync(
      `powershell -Command "Expand-Archive -Path '${archive}' -DestinationPath '${dest}' -Force"`,
      { stdio: 'inherit' }
    );
  } else {
    // Use tar on Unix
    execSync(`tar -xzf "${archive}" -C "${dest}"`, { stdio: 'inherit' });
  }
}

async function main() {
  try {
    console.log(`\n📦 Installing KothaSet v${VERSION}...\n`);

    // Create bin directory
    if (!fs.existsSync(BIN_DIR)) {
      fs.mkdirSync(BIN_DIR, { recursive: true });
    }

    const { platform } = getPlatformArch();
    const ext = platform === 'windows' ? 'zip' : 'tar.gz';
    const archivePath = path.join(BIN_DIR, `kothaset.${ext}`);

    // Download archive
    const url = getDownloadUrl();
    await download(url, archivePath);

    // Extract
    extract(archivePath, BIN_DIR);

    // Clean up archive
    fs.unlinkSync(archivePath);

    // Make binary executable (Unix only)
    const binaryPath = path.join(BIN_DIR, BINARY_NAME);
    if (platform !== 'windows') {
      fs.chmodSync(binaryPath, 0o755);
    }

    // Verify installation
    console.log('\n✅ KothaSet installed successfully!');
    console.log(`   Binary: ${binaryPath}`);
    console.log('\nRun "kothaset --help" to get started.\n');

  } catch (error) {
    console.error('\n❌ Installation failed:', error.message);
    console.error('\nManual installation:');
    console.error(`  1. Download from: https://github.com/${REPO}/releases`);
    console.error('  2. Extract and add to your PATH');
    process.exit(1);
  }
}

main();
