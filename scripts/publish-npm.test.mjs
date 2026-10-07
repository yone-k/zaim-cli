import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { test } from 'node:test';

function runPublish(t, response, args = []) {
  const dir = mkdtempSync(join(tmpdir(), 'zaim-publish-test-'));
  t.after(() => rmSync(dir, { recursive: true }));
  const bin = join(dir, 'bin');
  mkdirSync(bin);
  writeFileSync(join(bin, 'npm'), '#!/usr/bin/env bash\n' + response, { mode: 0o755 });
  const env = { ...process.env, PATH: `${bin}:${process.env.PATH}`, ACTIONS_ID_TOKEN_REQUEST_URL: 'https://actions.test/token', ACTIONS_ID_TOKEN_REQUEST_TOKEN: 'fake-request-token' };
  delete env.NODE_AUTH_TOKEN;
  delete env.NPM_TOKEN;
  return spawnSync('bash', ['scripts/publish-npm.sh', 'v0.3.0', ...args], { encoding: 'utf8', env });
}

test('publishes with OIDC without a stored npm token', t => {
  const result = runPublish(t, 'test "$1" = publish || exit 2\ntest "$(jq -r .version package.json)" = 0.3.0 || exit 3\necho published\n');
  assert.equal(result.status, 0, result.stderr + result.stdout);
  assert.match(result.stdout, /Published successfully/);
});

test('authentication failure remains a failure', t => {
  const result = runPublish(t, 'echo "npm error code E404: permission denied" >&2\nexit 1\n');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /E404: permission denied/);
  assert.doesNotMatch(result.stdout, /successfully|Already published/);
});

test('an already published version is safe to retry', t => {
  const result = runPublish(t, 'echo "npm error You cannot publish over the previously published versions" >&2\nexit 1\n');
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Already published/);
});

test('dry run does not invoke npm publish', t => {
  const result = runPublish(t, 'echo "unexpected npm invocation" >&2\nexit 99\n', ['--dry-run']);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /Dry run complete/);
});
