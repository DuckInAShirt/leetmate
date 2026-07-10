'use strict';

const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const fs = require('node:fs');
const os = require('node:os');
const test = require('node:test');
const path = require('node:path');
const { Readable } = require('node:stream');
const {
  assetName,
  binaryPath,
  download,
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

test('download follows redirects before writing destination', async (t) => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'leetmate-download-'));
  t.after(() => fs.rmSync(dir, { recursive: true, force: true }));

  const dest = path.join(dir, 'file.txt');
  const seen = [];
  const get = (url, _options, callback) => {
    seen.push(url);
    const request = new EventEmitter();
    process.nextTick(() => {
      if (url === 'https://example.test/start') {
        const response = new EventEmitter();
        response.statusCode = 302;
        response.headers = { location: '/final' };
        response.resume = () => {};
        callback(response);
        return;
      }
      if (url === 'https://example.test/final') {
        const response = Readable.from(['ok']);
        response.statusCode = 200;
        response.headers = {};
        callback(response);
        return;
      }
      request.emit('error', new Error(`unexpected url: ${url}`));
    });
    return request;
  };

  await download('https://example.test/start', dest, get);

  assert.deepEqual(seen, ['https://example.test/start', 'https://example.test/final']);
  assert.equal(fs.readFileSync(dest, 'utf8'), 'ok');
});
