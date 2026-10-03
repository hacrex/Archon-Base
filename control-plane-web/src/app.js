import { state } from './state.js';
import { modal } from './components.js';
import { overview, projects, resources, operations, admin, audit, settings } from './routes.js';

export function boot() {
  const app = document.querySelector('#app');
  const crumb = document.querySelector('#page-crumb');
  const pages = { overview: () => overview(state), projects, resources, operations, admin, audit, settings };

  function render() {
    const raw = location.hash.replace(/^#\/?/, '') || 'overview';
    const parts = raw.split('/');
    state.route = parts[0];
    state.subroute = parts.slice(1).join('/');
    app.innerHTML = (pages[state.route] || pages.overview)();
    crumb.textContent = state.subroute === 'allocation' ? 'Resource allocation' : state.route === 'overview' ? 'Overview' : state.route[0].toUpperCase() + state.route.slice(1);
    document.querySelectorAll('[data-route]').forEach((link) => link.classList.toggle('active', link.dataset.route === state.route));
    app.focus();
  }

  document.addEventListener('click', (event) => {
    const target = event.target.closest('[data-action]');
    if (!target) return;
    const action = target.dataset.action;
    if (action === 'dismiss-notice') document.querySelector('.notice-bar')?.remove();
    if (action === 'menu') document.querySelector('.sidebar').classList.toggle('open');
    if (action === 'context' || action === 'user') modal('Context switcher', 'The local preview currently has one organization and one development project. Organization membership and multi-project switching arrive with authentication.');
    if (action === 'search' || action === 'notifications') modal(action === 'search' ? 'Global search' : 'Notifications', action === 'search' ? 'Search will span projects, resources, deployments, logs, and documentation once the API index is connected.' : 'No new notifications. Operation events will appear here when the event stream is connected.');
    if (action === 'planned') modal('Planned capability', 'This control-plane surface is visible to make the product boundary clear. The backend contract and authorization flow are not implemented yet.');
    if (action === 'new-project' || action === 'new-resource') modal(action === 'new-project' ? 'Create project' : 'Create resource', 'Creation forms will be backed by the v1 API in the next slice. This preview does not submit changes.', '<button class="button" data-action="close-modal">Cancel</button><button class="button primary" data-action="planned">View status</button>');
    if (action === 'allocation-policy') modal('Allocation policy', 'Requests reserve capacity on the selected data plane. Limits cap runtime usage. Validation will be performed by the resource API when connected.');
    if (action === 'edit-allocation') { const item = state.allocations[Number(target.dataset.index || 0)]; modal('Edit allocation', `<div class="allocation-form"><label>CPU request<input value="${item.requested.cpu}" disabled></label><label>CPU limit<input value="${item.limit.cpu}" disabled></label><label>Memory request<input value="${item.requested.memory}" disabled></label><label>Memory limit<input value="${item.limit.memory}" disabled></label><label>GPU request<input value="${item.requested.gpu}" disabled></label></div><p class="form-note">The editor is staged for the resource API. Values are read-only in this preview.</p>`, '<button class="button primary" data-action="close-modal">Close</button>'); }
    if (action === 'close-modal') document.querySelector('#modal-root').innerHTML = '';
  });
  window.addEventListener('hashchange', render);
  render();
}
