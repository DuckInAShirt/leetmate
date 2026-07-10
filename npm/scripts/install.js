#!/usr/bin/env node
'use strict';

const crypto = require('crypto');
const fs = require('fs');
const https = require('https');
const os = require('os');
const path = require('path');
const { execFileSync } = require('child_process');

const root = path.resolve(__dirname, '..');
const binDir = path.join(root, 'bin');
const packageJSON = require(path.join(root, 'package.json'));
const owner = 'DuckInAShirt';
const repo = 'leetmate';

function platformTarget(platform = os.platform(), arch = os.arch()) {
  const osName = platform === 'darwin' ? 'macOS' : platform === 'win32' ? 'Windows' : platform === 'linux' ? 'linux' : '';
  const archName = arch === 'x64' ? 'amd64' : arch === 'arm64' ? 'arm64' : '';
  if (!osName || !archName) {
    throw new Error(`Unsupported platform: ${platform}/${arch}`);
  }
  return {
    osName,
    archName,
    ext: platform === 'win32' ? '.zip' : '.tar.gz',
    exe: platform === 'win32' ? 'leetmate.exe' : 'leetmate',
  };
}

function releaseVersion(env = process.env, pkg = packageJSON) {
  if (env.LEETMATE_VERSION) {
    return env.LEETMATE_VERSION.replace(/^v/, '');
  }
  return pkg.version.replace(/^v/, '');
}

function assetName(version, target) {
  return `leetmate_${version}_${target.osName}_${target.archName}${target.ext}`;
}

function releaseURL(version, name) {
  return `https://github.com/${owner}/${repo}/releases/download/v${version}/${name}`;
}

function binaryPath(target) {
  return path.join(binDir, target.exe);
}

function parseChecksums(text) {
  const out = new Map();
  for (const line of text.split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const match = trimmed.match(/^([a-fA-F0-9]{64})\s+(.+)$/);
    if (match) {
      out.set(match[2].replace(/^\*?/, ''), match[1].toLowerCase());
    }
  }
  return out;
}

function sha256(file) {
  const hash = crypto.createHash('sha256');
  hash.update(fs.readFileSync(file));
  return hash.digest('hex');
}

function download(url, dest, get = https.get, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) {
      reject(new Error(`Too many redirects while downloading ${url}`));
      return;
    }

    const request = get(url, { headers: { 'User-Agent': 'leetmate-npm-installer' } }, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        const nextURL = new URL(response.headers.location, url).toString();
        download(nextURL, dest, get, redirects + 1).then(resolve, reject);
        return;
      }
      if (response.statusCode !== 200) {
        response.resume();
        fs.rmSync(dest, { force: true });
        reject(new Error(`Download failed: HTTP ${response.statusCode} ${url}`));
        return;
      }

      const file = fs.createWriteStream(dest);
      const cleanup = (err) => {
        file.close(() => {
          fs.rmSync(dest, { force: true });
          reject(err);
        });
      };
      response.on('error', cleanup);
      file.on('error', cleanup);
      file.on('finish', () => file.close(resolve));
      response.pipe(file);
    });
    request.on('error', reject);
  });
}

function extract(archive, target) {
  fs.mkdirSync(binDir, { recursive: true });
  const binary = binaryPath(target);
  fs.rmSync(binary, { force: true });
  if (archive.endsWith('.zip')) {
    const tmpDir = `${archive}-extract`;
    fs.rmSync(tmpDir, { recursive: true, force: true });
    fs.mkdirSync(tmpDir, { recursive: true });
    execFileSync('powershell', ['-NoProfile', '-Command', `Expand-Archive -LiteralPath ${JSON.stringify(archive)} -DestinationPath ${JSON.stringify(tmpDir)} -Force`], { stdio: 'inherit' });
    fs.copyFileSync(path.join(tmpDir, target.exe), binary);
    fs.rmSync(tmpDir, { recursive: true, force: true });
  } else {
    execFileSync('tar', ['-xzf', archive, '-C', binDir, target.exe], { stdio: 'inherit' });
  }
  if (process.platform !== 'win32') {
    fs.chmodSync(binary, 0o755);
  }
}

async function install() {
  const target = platformTarget();
  const skipDownload = process.argv.includes('--skip-download') || process.env.LEETMATE_SKIP_DOWNLOAD === '1';
  if (skipDownload) {
    fs.mkdirSync(binDir, { recursive: true });
    const binary = binaryPath(target);
    if (!fs.existsSync(binary)) {
      fs.writeFileSync(binary, '#!/bin/sh\necho leetmate development\n', { mode: 0o755 });
    }
    return;
  }

  const version = releaseVersion();
  if (!version || version === '0.0.0-dev') {
    throw new Error('Cannot install development npm package without LEETMATE_VERSION. Use a published release package.');
  }

  const name = assetName(version, target);
  const tmp = path.join(os.tmpdir(), `${name}-${process.pid}`);
  const sums = path.join(os.tmpdir(), `leetmate-checksums-${version}-${process.pid}.txt`);
  await download(releaseURL(version, name), tmp);
  await download(releaseURL(version, 'checksums.txt'), sums);
  const checksums = parseChecksums(fs.readFileSync(sums, 'utf8'));
  const expected = checksums.get(name);
  if (!expected) {
    throw new Error(`No checksum found for ${name}`);
  }
  const actual = sha256(tmp);
  if (actual !== expected) {
    throw new Error(`Checksum mismatch for ${name}: expected ${expected}, got ${actual}`);
  }
  extract(tmp, target);
  fs.rmSync(tmp, { force: true });
  fs.rmSync(sums, { force: true });
}

if (require.main === module) {
  install().catch((err) => {
    console.error(`leetmate install failed: ${err.message}`);
    process.exit(1);
  });
}

module.exports = { assetName, binaryPath, download, parseChecksums, platformTarget, releaseURL, releaseVersion };
