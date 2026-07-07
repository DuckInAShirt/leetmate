#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const exe = process.platform === 'win32' ? 'leetmate.exe' : 'leetmate';
const binary = path.join(__dirname, exe);

if (!fs.existsSync(binary)) {
  console.error('leetmate binary is missing. Reinstall the package or run `npm rebuild leetmate`.');
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
