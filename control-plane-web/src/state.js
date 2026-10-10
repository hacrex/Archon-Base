export const state = {
  token: sessionStorage.getItem('archon_token') || '',
  user: null,
  authLoading: false,
  authError: '',
  projects: [],
  projectsLoading: false,
  projectsError: '',
  route: 'overview',
  subroute: '',
  resources: [
    { name: 'support-bot', type: 'Project', status: 'Ready', detail: 'development · local' },
    { name: 'docs-index', type: 'DatabaseInstance', status: 'Ready', detail: 'qdrant · standard-2' },
    { name: 'support-agent', type: 'AgentFunction', status: 'Planned', detail: 'python3.12 · sandbox' },
  ],
  allocations: [
    { resource: 'docs-index', kind: 'DatabaseInstance', cpu: '0.50 / 1.00', memory: '1 / 2 GiB', storage: '20 / 50 GiB', gpu: '—', requested: { cpu: '500m', memory: '1Gi', gpu: '0' }, limit: { cpu: '1', memory: '2Gi', gpu: '0' } },
    { resource: 'support-agent', kind: 'AgentFunction', cpu: 'Planned', memory: 'Planned', storage: 'Planned', gpu: 'Planned', requested: { cpu: '250m', memory: '512Mi', gpu: '0' }, limit: { cpu: '1', memory: '1Gi', gpu: '0' } },
  ],
};
