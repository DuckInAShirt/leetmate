'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');
const path = require('node:path');
const {
  assetName,
  binaryPath,
  parseChecksums,
  platformTarget,
  releaseURL,
  releaseVersion,
} = require('./install');

test('platformTarget maps supported platforms to GoReleaser names', () => {
  assert.deepEqual(platformTarget('darwin', 'arm64'), { osName: 'macOS', archName: 'arm64', ext: '.tar.gz', exe: 'leetmate' });
  assert.deepEqual(platformTarget('linux', 'x64'), { osName: 'linux', archName: 'amd64', ext: '.tar.gz', exe: 'leetmate' });
  assert.deepEqual(platformTarget('win32', 'x64'), { osName: 'Windows', archName: 'amd64', ext: '.zip', exe: 'leetmate.exe' });
});

test('platformTarget rejects unsupported combinations', () => {
  assert.throws(() => platformTarget('freebsd', 'x64'), /Unsupported platform/);
  assert.throws(() => platformTarget('linux', 'arm'), /Unsupported platform/);
});

test('assetName and releaseURL match GoReleaser assets', () => {
  const target = platformTarget('darwin', 'arm64');
  const name = assetName('0.2.7', target);
  assert.equal(name, 'leetmate_0.2.7_macOS_arm64.tar.gz');
  assert.equal(releaseURL('0.2.7', name), 'https://github.com/DuckInAShirt/leetmate/releases/download/v0.2.7/leetmate_0.2.7_macOS_arm64.tar.gz');
});

test('parseChecksums extracts sha256 entries', () => {
  const sum = 'a'.repeat(64);
  const parsed = parseChecksums(`${sum}  leetmate_0.2.7_linux_amd64.tar.gz\n`);
  assert.equal(parsed.get('leetmate_0.2.7_linux_amd64.tar.gz'), sum);
});

test('releaseVersion prefers LEETMATE_VERSION', () => {
  assert.equal(releaseVersion({ LEETMATE_VERSION: 'v1.2.3' }, { version: '0.0.0-dev' }), '1.2.3');
  assert.equal(releaseVersion({}, { version: '1.2.4' }), '1.2.4');
});

test('binaryPath points inside npm bin directory', () => {
  assert.equal(path.basename(binaryPath(platformTarget('win32', 'x64'))), 'leetmate.exe');
  assert.equal(path.basename(binaryPath(platformTarget('linux', 'x64'))), 'leetmate');
});
