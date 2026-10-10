import { state } from './state.js';
import { login, me, logout, listProjects, createProject } from './api.js';
import { modal } from './components.js';
import { overview, projects, resources, operations, admin, audit, settings } from './routes.js';
import { loginScreen, liveProjects } from './live-routes.js';

export function boot() {
  const app = document.querySelector('#app');
  const crumb = document.querySelector('#page-crumb');
  const pages = { overview: () => overview(state), projects: () => liveProjects(state), resources, operations, admin, audit, settings };

  function render() {
    document.body.classList.toggle('auth-mode', !state.token);
    if (!state.token) {
      app.innerHTML = loginScreen(state);
      return;
    }
    const raw = location.hash.replace(/^#\/?/, '') || 'overview';
    const parts = raw.split('/');
    state.route = parts[0];
    state.subroute = parts.slice(1).join('/');
    app.innerHTML = (pages[state.route] || pages.overview)();
    crumb.textContent = state.subroute === 'allocation' ? 'Resource allocation' : state.route === 'overview' ? 'Overview' : state.route[0].toUpperCase() + state.route.slice(1);
    document.querySelectorAll('[data-route]').forEach((link) => link.classList.toggle('active', link.dataset.route === state.route));
    app.focus();
  }

  async function loadProjects() {
    state.projectsLoading = true;
    render();
    try {
      state.projects = (await listProjects(state.token)).items || [];
      state.projectsError = '';
    } catch (error) {
      if (error.status === 401) {
        state.token = '';
        sessionStorage.removeItem('archon_token');
      }
      state.projectsError = error.message;
    } finally {
      state.projectsLoading = false;
    }
  }

  async function loadSession() {
    if (!state.token) return render();
    try {
      state.user = await me(state.token);
      await loadProjects();
    } catch (error) {
      state.token = '';
      sessionStorage.removeItem('archon_token');
      state.authError = error.status === 401 ? 'Your session expired. Please sign in again.' : error.message;
    }
    render();
  }

  document.addEventListener('submit', async (event) => {
    const form = event.target.closest('[data-form]');
    if (!form) return;
    event.preventDefault();
    const values = Object.fromEntries(new FormData(form));
    if (form.dataset.form === 'login') {
      state.authLoading = true;
      state.authError = '';
      render();
      try {
        const result = await login(values.email, values.password);
        state.token = result.token;
        sessionStorage.setItem('archon_token', result.token);
        state.authLoading = false;
        await loadSession();
      } catch (error) {
        state.authLoading = false;
        state.authError = error.status === 401 ? 'Email or password is incorrect.' : error.message;
        render();
      }
    }
    if (form.dataset.form === 'project') {
      const submit = form.querySelector('button[type="submit"]');
      submit.disabled = true;
      try {
        await createProject(state.token, values);
        document.querySelector('#modal-root').innerHTML = '';
        location.hash = '#/projects';
        await loadProjects();
      } catch (error) {
        const message = form.querySelector('.form-error');
        message.hidden = false;
        message.textContent = error.message;
        submit.disabled = false;
      }
    }
  });

  document.addEventListener('click', async (event) => {
    const target = event.target.closest('[data-action]');
    if (!target) return;
    const action = target.dataset.action;
    if (action === 'dismiss-notice') document.querySelector('.notice-bar')?.remove();
    if (action === 'menu') document.querySelector('.sidebar').classList.toggle('open');
    if (action === 'user') {
      modal('Account', `<p>${state.user?.Email || 'Signed-in user'}</p><p class="form-note">Organization: ${state.user?.Organization || 'local'}</p>`, '<button class="button" data-action="close-modal">Cancel</button><button class="button primary" data-action="logout">Sign out</button>');
    }
    if (action === 'context') modal('AI Backend context', 'Organization and project switching will use authenticated membership and project APIs in the next slice.');
    if (action === 'search' || action === 'notifications') modal(action === 'search' ? 'Global search' : 'Notifications', action === 'search' ? 'Search will span projects, resources, deployments, logs, and documentation once the API index is connected.' : 'No new notifications.');
    if (action === 'new-project') {
      modal('Create AI Backend project', `<form data-form="project" class="auth-form"><label>Project name<input name="name" required pattern="[a-z][a-z0-9-]{1,62}" placeholder="support-bot"></label><label>Display name<input name="displayName" required placeholder="Support Bot"></label><label>Environment<select name="environment"><option>development</option><option>staging</option><option>production</option></select></label><label>Region<input name="region" value="local"></label><div class="form-error" hidden></div><button class="button primary" type="submit">Create project</button></form>`);
    }
    if (action === 'logout') {
      try { await logout(state.token); } catch (_) { /* clear local state even if API is unavailable */ }
      state.token = '';
      state.user = null;
      sessionStorage.removeItem('archon_token');
      document.querySelector('#modal-root').innerHTML = '';
      render();
    }
    if (action === 'planned') modal('Planned capability', 'This AI Backend surface is visible to make the product boundary clear.');
    if (action === 'allocation-policy' || action === 'edit-allocation') modal('Preview capability', 'Allocation management will be connected to the resource API after project authorization is complete.');
    if (action === 'close-modal') document.querySelector('#modal-root').innerHTML = '';
  });
  window.addEventListener('hashchange', render);
  loadSession();
}
