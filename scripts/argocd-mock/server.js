// Minimal ArgoCD API mock for local testing of the Release Portal.
// Implements exactly the endpoints used by backend/internal/integration/argocd/client.go:
//   GET  /api/v1/session/userinfo          (TestConnection)
//   GET  /api/v1/applications              (ListApplications)
//   GET  /api/v1/applications/:name        (GetApplication)
//   POST /api/v1/applications/:name/sync   (SyncApplication)
//   PUT  /api/v1/applications/:name        (patchApplication - JSON patch array)
// State is kept in memory; annotations/spec changes persist until restart.
const http = require('http');

const PORT = process.env.PORT || 8080;

let seq = 1;
const uid = () => `mock-uid-${seq++}`;

function app(name, ns, repo, path, revision, extraAnnotations = {}) {
  return {
    metadata: {
      name,
      namespace: ns,
      uid: uid(),
      resourceVersion: String(seq),
      annotations: {
        'canary.argoproj.io/traffic-split': '0',
        'canary.argoproj.io/mode': 'stable',
        ...extraAnnotations,
      },
    },
    spec: {
      source: { repoURL: repo, path, targetRevision: revision },
      destination: { server: 'https://kubernetes.default.svc', namespace: ns },
    },
    status: {
      sync: { status: 'Synced', revision: `${revision}-sha${seq}` },
      health: { status: 'Healthy' },
    },
  };
}

const appsByName = new Map();
[
  ['payments-api', 'production', 'http://gitlab.local/root/payments-api.git', 'deploy/argocd', 'v1.2.0'],
  ['auth-service', 'production', 'http://gitlab.local/root/auth-service.git', 'deploy/argocd', 'v0.9.3'],
  ['web-frontend', 'staging', 'http://gitlab.local/root/web-frontend.git', 'deploy/argocd', 'main'],
].forEach(([n, ns, repo, p, rev]) => appsByName.set(n, app(n, ns, repo, p, rev)));

function json(res, code, obj) {
  res.writeHead(code, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(obj));
}

function readBody(req) {
  return new Promise((resolve) => {
    let data = '';
    req.on('data', (c) => (data += c));
    req.on('end', () => {
      try { resolve(data ? JSON.parse(data) : null); } catch { resolve(null); }
    });
  });
}

// RFC6902-ish JSON patch supporting add/replace/remove on paths like
// /spec/source/targetRevision and /metadata/annotations/canary.argoproj.io~1traffic-split
function applyPatch(target, patches) {
  for (const p of patches) {
    const parts = p.path.replace(/^\/+/, '').split('/').map((s) => s.replace(/~1/g, '/').replace(/~0/g, '~'));
    let obj = target;
    for (let i = 0; i < parts.length - 1; i++) {
      if (obj[parts[i]] === undefined) obj[parts[i]] = {};
      obj = obj[parts[i]];
    }
    const last = parts[parts.length - 1];
    if (p.op === 'remove') delete obj[last];
    else obj[last] = p.value; // add | replace
  }
}

const server = http.createServer(async (req, res) => {
  const url = new URL(req.url, `http://${req.headers.host}`);
  const seg = url.pathname.split('/').filter(Boolean); // e.g. api/v1/applications/<name>/sync

  console.log(`[argocd-mock] ${req.method} ${url.pathname}`);

  // /api/v1/session/userinfo -> TestConnection
  if (req.method === 'GET' && url.pathname === '/api/v1/session/userinfo') {
    return json(res, 200, { sub: 'admin', exp: Math.floor(Date.now() / 1000) + 3600, iss: 'argocd-uid', username: 'admin', full_name: 'Administrator', groups: [], sessToken: true });
  }

  if (seg[0] === 'api' && seg[1] === 'v1' && seg[2] === 'applications') {
    const name = decodeURIComponent(seg[3] || '');

    if (!name && req.method === 'GET') {
      return json(res, 200, { items: [...appsByName.values()] });
    }

    if (name && req.method === 'GET' && seg.length === 4) {
      const a = appsByName.get(name);
      if (!a) return json(res, 404, { error: { message: `applications.argoproj.io "${name}" not found`, code: 5 } });
      return json(res, 200, a);
    }

    if (name && seg[4] === 'sync' && req.method === 'POST') {
      const a = appsByName.get(name);
      if (!a) return json(res, 404, { error: { message: `applications.argoproj.io "${name}" not found`, code: 5 } });
      a.status.sync.status = 'Synced';
      a.status.health.status = 'Healthy';
      return json(res, 200, a);
    }

    if (name && req.method === 'PUT') {
      const a = appsByName.get(name);
      if (!a) return json(res, 404, { error: { message: `applications.argoproj.io "${name}" not found`, code: 5 } });
      const body = await readBody(req);
      if (Array.isArray(body)) {
        applyPatch(a, body);
      } else if (body && body.spec) {
        a.spec = body.spec;
      }
      a.metadata.resourceVersion = String(seq++);
      return json(res, 200, a);
    }
  }

  json(res, 404, { error: { message: `not found: ${req.method} ${url.pathname}`, code: 12 } });
});

server.listen(PORT, () => console.log(`[argocd-mock] listening on :${PORT}`));
