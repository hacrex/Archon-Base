const API_BASE = window.ARCHON_API_URL || 'http://127.0.0.1:8080';

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const error = new Error(body.message || `API request failed (${response.status})`);
    error.status = response.status;
    error.body = body;
    throw error;
  }
  return body;
}

export function login(email, password) {
  return request('/v1/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
}

export function me(token) {
  return request('/v1/auth/me', { headers: { Authorization: `Bearer ${token}` } });
}

export function logout(token) {
  return request('/v1/auth/logout', { method: 'POST', headers: { Authorization: `Bearer ${token}` } });
}

export function listProjects(token) {
  return request('/v1/projects', { headers: { Authorization: `Bearer ${token}` } });
}

export function createProject(token, project) {
  return request('/v1/projects', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify({ metadata: { name: project.name }, spec: { displayName: project.displayName, description: project.description, environment: project.environment, region: project.region } }) });
}
