import assert from 'node:assert/strict';
import { test } from 'node:test';
import { verifyOIDC } from './verify-npm-oidc.mjs';

const env = { ACTIONS_ID_TOKEN_REQUEST_URL: 'https://actions.test/token?existing=1', ACTIONS_ID_TOKEN_REQUEST_TOKEN: 'request-secret' };
const issued = { token_type: 'oidc', token: 'npm-secret', expires: '2099-01-01T00:00:00Z' };

test('exchanges a GitHub identity for the exact package without exposing tokens', async () => {
  const calls = [];
  const fetchImpl = async (url, options) => {
    calls.push({ url: String(url), options });
    return calls.length === 1 ? Response.json({ value: 'github-secret' }) : Response.json(issued, { status: 201 });
  };
  const result = await verifyOIDC('@yone_k/zaim-cli', { env, fetchImpl });
  assert.equal(calls.length, 2);
  const requestURL = new URL(calls[0].url);
  assert.equal(requestURL.searchParams.get('audience'), 'npm:registry.npmjs.org');
  assert.equal(calls[0].options.headers.Authorization, 'Bearer request-secret');
  assert.equal(calls[1].url, 'https://registry.npmjs.org/-/npm/v1/oidc/token/exchange/package/%40yone_k%2Fzaim-cli');
  assert.equal(calls[1].options.method, 'POST');
  assert.equal(calls[1].options.headers.Authorization, 'Bearer github-secret');
  assert.deepEqual(result, { packageName: '@yone_k/zaim-cli', expires: issued.expires });
});

test('missing OIDC credentials fail before any request', async () => {
  await assert.rejects(verifyOIDC('@yone_k/zaim-cli', { env: {}, fetchImpl: () => assert.fail('unexpected request') }), /OIDC environment/);
});

test('a rejected GitHub identity fails without an npm request', async () => {
  let calls = 0;
  await assert.rejects(verifyOIDC('@yone_k/zaim-cli', { env, fetchImpl: async () => { calls++; return new Response('secret-body', { status: 403 }); } }), /GitHub OIDC.*403/);
  assert.equal(calls, 1);
});

test('a denied npm identity remains an error without leaking the response', async () => {
  let calls = 0;
  await assert.rejects(verifyOIDC('@yone_k/zaim-cli', { env, fetchImpl: async () => ++calls === 1 ? Response.json({ value: 'github-secret' }) : new Response('npm-secret', { status: 404 }) }), error => /npm OIDC.*404/.test(error.message) && !error.message.includes('npm-secret'));
});

test('successful HTTP with no issued token does not pass', async () => {
  let calls = 0;
  await assert.rejects(verifyOIDC('@yone_k/zaim-cli', { env, fetchImpl: async () => ++calls === 1 ? Response.json({ value: 'github-secret' }) : Response.json({ token_type: 'oidc' }, { status: 201 }) }), /valid publishing token/);
});
