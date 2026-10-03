export function pageHead(eyebrow, title, description, action = '') {
  const label = action === 'new-project' ? 'New project' : 'Create resource';
  return `<div class="page-heading"><div><div class="eyebrow">${eyebrow}</div><h1>${title}</h1><p>${description}</p></div>${action ? `<button class="button primary" data-action="${action}">＋ ${label}</button>` : ''}</div>`;
}

export function status(value) {
  const cls = value.toLowerCase() === 'ready' ? 'ready' : value.toLowerCase() === 'planned' ? 'planned' : 'pending';
  return `<span class="status-badge ${cls}">${value}</span>`;
}

export function allocationBar(label, used, total, unit) {
  const pct = Math.min(100, Math.round((used / total) * 100));
  return `<div class="allocation-line"><div class="allocation-line-top"><span>${label}</span><b>${used} / ${total} ${unit}</b></div><div class="meter"><span style="width:${pct}%"></span></div><div class="allocation-line-foot"><span>${pct}% allocated</span><span>${Math.max(0, total - used)} ${unit} available</span></div></div>`;
}

export function modal(title, body, actions = '') {
  document.querySelector('#modal-root').innerHTML = `<div class="modal-backdrop" data-action="close-modal"><div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title" onclick="event.stopPropagation()"><h2 id="modal-title">${title}</h2><div class="modal-body">${body}</div><div class="modal-actions">${actions || '<button class="button primary" data-action="close-modal">Close</button>'}</div></div></div>`;
}
