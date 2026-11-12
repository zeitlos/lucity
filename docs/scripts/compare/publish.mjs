import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';

const bucket = process.env.COMPARE_BUCKET || 'zeitlos-software/lucity/production/media';
const platformURL = (process.env.COMPARE_LUCITY_URL || 'https://lucity.cloud').replace(/\/$/, '');
const workspace = bucket.split('/')[0];

let token;

function session() {
  if (!token) {
    try {
      token = execFileSync('lucity', ['token'], {
        encoding: 'utf8',
        stdio: ['ignore', 'pipe', 'pipe'],
        env: { ...process.env, LUCITY_WORKSPACE: workspace },
      }).trim();
    } catch (error) {
      throw new Error(`lucity token failed, sign the CLI in to ${platformURL}: ${error.stderr?.toString().trim() || error.message}`);
    }
  }
  return token;
}

async function graphql(query, variables = {}) {
  const response = await fetch(`${platformURL}/graphql`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${session()}`, 'X-Lucity-Workspace': workspace },
    body: JSON.stringify({ query, variables }),
  });
  const payload = await response.json().catch(() => ({ errors: [{ message: `${response.status} from ${platformURL}` }] }));
  if (payload.errors) throw new Error(`${payload.errors.map((error) => error.message).join(', ')} (is the CLI signed in to ${platformURL}?)`);
  return payload.data;
}

/** A short content hash, so a changed file gets a new name and never hides behind a cached old one. */
export function fingerprint(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex').slice(0, 10);
}

/**
 * Uploads a file to the public media bucket under `key` and returns its public
 * URL. Uses the Lucity CLI's session; no bucket keys are stored anywhere.
 */
export async function publish(file, key, contentType) {
  const { bucket: { publicEndpoint, public: open } } = await graphql(
    'query($bucket: BucketID!) { bucket(id: $bucket) { publicEndpoint public } }',
    { bucket },
  );
  if (!open || !publicEndpoint) throw new Error(`${bucket} is not public, so its files have no public URL`);
  const { bucketObjectUploadUrl: uploadURL } = await graphql(
    'mutation($bucket: BucketID!, $key: String!) { bucketObjectUploadUrl(bucket: $bucket, key: $key) }',
    { bucket, key },
  );
  const response = await fetch(uploadURL, { method: 'PUT', headers: { 'Content-Type': contentType }, body: readFileSync(file) });
  if (!response.ok) throw new Error(`uploading ${key} failed: ${response.status} ${await response.text()}`);
  return `${publicEndpoint.replace(/\/$/, '')}/${key}`;
}
