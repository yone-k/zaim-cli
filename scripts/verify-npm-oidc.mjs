import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

// Verify CI's publishing identity without publishing a new package version.
export async function verifyOIDC(packageName, { env = process.env, fetchImpl = fetch } = {}) {
  const requestURL = env.ACTIONS_ID_TOKEN_REQUEST_URL;
  const requestToken = env.ACTIONS_ID_TOKEN_REQUEST_TOKEN;
  if (!requestURL || !requestToken) throw new Error('GitHub OIDC environment is missing');
  const url = new URL(requestURL);
  url.searchParams.set('audience', 'npm:registry.npmjs.org');
  const identityResponse = await fetchImpl(url, { headers: { Authorization: `Bearer ${requestToken}` } });
  if (!identityResponse.ok) throw new Error(`GitHub OIDC request failed (HTTP ${identityResponse.status})`);
  const identity = await identityResponse.json();
  if (typeof identity.value !== 'string' || !identity.value) throw new Error('GitHub did not issue an OIDC identity');
  const exchangeResponse = await fetchImpl(`https://registry.npmjs.org/-/npm/v1/oidc/token/exchange/package/${encodeURIComponent(packageName)}`, {
    method: 'POST', headers: { Authorization: `Bearer ${identity.value}` },
  });
  if (exchangeResponse.status !== 201) throw new Error(`npm OIDC exchange failed (HTTP ${exchangeResponse.status})`);
  const issued = await exchangeResponse.json();
  // npm CLI consumes response.token; token_type and expires are optional metadata.
  if (typeof issued.token !== 'string' || !issued.token) {
    throw new Error(`npm did not issue a valid publishing token (response fields: ${Object.keys(issued).join(', ')})`);
  }
  // The issued credential is deliberately excluded from output and logs.
  const expires = Number.isFinite(Date.parse(issued.expires)) ? issued.expires : undefined;
  return { packageName, expires };
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const verified = await verifyOIDC(process.argv[2] || '@yone_k/zaim-cli');
    console.log(`npm OIDC publishing identity verified for ${verified.packageName}${verified.expires ? ` (expires ${verified.expires})` : ''}`);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
