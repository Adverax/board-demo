// Board page (COMP-003, SCR-001). Plain JavaScript, no build step (CON-005).
//
// One render path (ADR-0009): after every action, succeeded or refused, the
// page shows any reason in the RefusalNotice, then re-reads GET /api/board and
// re-renders from it. There is no polling (ASM-0004): the board changes only
// after the user's own action. At most one action is in flight per page.
'use strict';

(() => {
  // One bound per fetch; an action is at most two requests (mutation, re-read).
  const REQUEST_TIMEOUT_MS = 5000;
  // Column order (INV-005), used only to pick the neighbour of a move and to
  // disable the edge controls (ADR-0008).
  const COLUMNS = ['todo', 'doing', 'done'];
  // INV-002, used only for the WIP-limit outline; the server enforces it.
  const WIP_LIMIT = 3;

  // Reasons for the CON-004 codes (FR-011); the server sends only the code.
  const REASONS = {
    wip_limit_exceeded: 'doing already holds 3 cards (limit 3).',
    invalid_move: 'the card is no longer where this page showed it. The board below shows its current place.',
    not_found: 'the card no longer exists. Another member may have deleted it.',
    validation_failed: 'a title is required and can be at most 80 characters.',
  };

  const ACTIONS = [
    { action: 'left', label: 'Move left', icon: 'icon-left' },
    { action: 'right', label: 'Move right', icon: 'icon-right' },
    { action: 'edit', label: 'Edit', icon: 'icon-edit' },
    { action: 'delete', label: 'Delete', icon: 'icon-delete' },
  ];

  const state = {
    board: { columns: [] }, // the last board read from the server
    busy: false,            // an action is in flight
    editing: null,          // {id, title, description, invalid} while a card is in edit mode
  };

  const $ = (sel) => document.querySelector(sel);
  const ui = {
    board: $('#board'),
    notice: $('#refusal'),
    addForm: $('#add-form'),
    addTitle: $('#add-title'),
    addDescription: $('#add-description'),
    addSubmit: $('#add-submit'),
  };

  // request performs one fetch bounded by REQUEST_TIMEOUT_MS and never
  // rejects. It resolves to {ok: true, body} or {ok: false, code, status}, where
  // code is a server code, 'network' (no response or timed out) or
  // 'unexpected' (a response that is not the API's shape).
  async function request(method, path, payload) {
    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), REQUEST_TIMEOUT_MS);
    try {
      const init = { method, signal: ctrl.signal, headers: { Accept: 'application/json' } };
      if (payload !== undefined) {
        init.headers['Content-Type'] = 'application/json';
        init.body = JSON.stringify(payload);
      }
      const resp = await fetch(path, init);
      // Read every body to its end, the empty one of a 204 too, so that no
      // response is left half-consumed.
      const body = await readJSON(resp);
      if (resp.ok) {
        // ADR-0009: a mutation's card is not used; the page renders from the re-read.
        return { ok: true, body };
      }
      const code = body && typeof body.code === 'string' && body.code !== '' ? body.code : 'unexpected';
      return { ok: false, code, status: resp.status };
    } catch (err) {
      return { ok: false, code: 'network', status: 0 };
    } finally {
      clearTimeout(timer);
    }
  }

  // readJSON returns the parsed body, or undefined when it is empty or not
  // JSON. An abort while reading is rethrown so that it counts as a network
  // failure.
  async function readJSON(resp) {
    const text = await resp.text();
    try {
      return text === '' ? undefined : JSON.parse(text);
    } catch (err) {
      return undefined;
    }
  }

  function isBoard(body) {
    return body !== null && typeof body === 'object' && Array.isArray(body.columns) &&
      body.columns.every((c) => c && typeof c.name === 'string' && Array.isArray(c.cards));
  }

  function reasonFor(outcome) {
    switch (outcome.code) {
      case 'network':
        return 'the board could not be reached. Check the connection and try again.';
      case 'unexpected':
        return `the server sent an unexpected response (HTTP ${outcome.status}).`;
      default:
        return REASONS[outcome.code] || `the server refused it (code "${outcome.code}").`;
    }
  }

  function showNotice(prefix, outcome) {
    ui.notice.textContent = `${prefix}: ${reasonFor(outcome)}`;
    ui.notice.dataset.code = outcome.code;
    ui.notice.hidden = false;
  }

  function hideNotice() {
    ui.notice.hidden = true;
    ui.notice.textContent = '';
    delete ui.notice.dataset.code;
  }

  function setBusy(busy) {
    state.busy = busy;
    ui.board.setAttribute('aria-busy', String(busy));
    ui.addSubmit.disabled = busy;
    const save = $('#edit-save');
    if (save) save.disabled = busy;
  }

  // refresh re-reads the board and re-renders it. On failure the last board
  // stays on the page and the notice says why.
  async function refresh(prefix) {
    const outcome = await request('GET', '/api/board');
    if (outcome.ok && isBoard(outcome.body)) {
      render(outcome.body);
      return;
    }
    showNotice(prefix, outcome.ok ? { code: 'unexpected', status: 200 } : outcome);
  }

  // act runs one action: the mutation, the reason (if refused), the re-read.
  // A call while another action is in flight is ignored: no request is sent.
  async function act(prefix, method, path, payload, onDone) {
    if (state.busy) return;
    setBusy(true);
    try {
      const outcome = await request(method, path, payload);
      if (outcome.ok) {
        hideNotice();
      } else {
        showNotice(prefix, outcome);
      }
      if (onDone) onDone(outcome);
      await refresh('Cannot reload the board');
    } finally {
      setBusy(false);
    }
  }

  function findCard(id) {
    for (const col of state.board.columns) {
      const card = col.cards.find((c) => c.id === id);
      if (card) return card;
    }
    return null;
  }

  function cardPath(id, suffix = '') {
    return `/api/cards/${encodeURIComponent(id)}${suffix}`;
  }

  // --- rendering -----------------------------------------------------------

  function el(tag, props = {}, children = []) {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(props)) {
      if (k === 'text') node.textContent = v;
      else if (k === 'class') node.className = v;
      else node.setAttribute(k, v);
    }
    node.append(...children);
    return node;
  }

  function icon(id) {
    const ns = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('aria-hidden', 'true');
    svg.setAttribute('focusable', 'false');
    const use = document.createElementNS(ns, 'use');
    use.setAttribute('href', `#${id}`);
    svg.append(use);
    return svg;
  }

  function cardView(card, columnIndex) {
    const body = [el('h3', { class: 'card-title', text: card.title })];
    if (card.description) {
      body.push(el('p', { class: 'card-description', text: card.description }));
    }
    const buttons = ACTIONS.map(({ action, label, icon: iconID }) => {
      const b = el('button', { type: 'button', class: 'icon-button', 'data-action': action, 'aria-label': label, title: label }, [icon(iconID)]);
      // ADR-0008: edge moves are rendered but disabled; no request is sent.
      if ((action === 'left' && columnIndex === 0) || (action === 'right' && columnIndex === COLUMNS.length - 1)) {
        b.disabled = true;
      }
      return b;
    });
    body.push(el('div', { class: 'card-actions' }, buttons));
    return el('li', { class: 'card', 'data-card-id': card.id }, body);
  }

  function editView(card, draft) {
    const title = el('input', { id: 'edit-title', name: 'title', 'aria-label': 'Title', autocomplete: 'off' });
    title.value = draft.title;
    if (draft.invalid) title.setAttribute('aria-invalid', 'true');
    const description = el('input', { id: 'edit-description', name: 'description', 'aria-label': 'Description', autocomplete: 'off', placeholder: 'Description (optional)' });
    description.value = draft.description;
    const form = el('form', { id: 'edit-card', class: 'edit-form', novalidate: '' }, [
      title,
      description,
      el('div', { class: 'edit-buttons' }, [
        el('button', { id: 'edit-save', type: 'submit', class: 'button-primary', text: 'Save' }),
        el('button', { type: 'button', class: 'button-secondary', 'data-action': 'cancel', text: 'Cancel' }),
      ]),
    ]);
    form.addEventListener('submit', (e) => {
      e.preventDefault();
      saveEdit(card, title.value, description.value);
    });
    return el('li', { class: 'card card-editing', 'data-card-id': card.id }, [form]);
  }

  // render draws the board exactly as the server sent it and keeps keyboard
  // focus on the same control of the same card when that card still exists.
  function render(board) {
    const focus = focusedControl();
    state.board = board;
    if (state.editing && !findCard(state.editing.id)) state.editing = null;
    COLUMNS.forEach((name, index) => {
      const section = ui.board.querySelector(`[data-column="${name}"]`);
      const column = board.columns.find((c) => c.name === name);
      const cards = column ? column.cards : [];
      section.querySelector('.count').textContent = String(cards.length);
      section.classList.toggle('column-at-limit', name === 'doing' && cards.length >= WIP_LIMIT);
      const items = cards.map((card) =>
        state.editing && state.editing.id === card.id ? editView(card, state.editing) : cardView(card, index));
      if (items.length === 0) items.push(el('li', { class: 'column-empty', text: 'No cards' }));
      section.querySelector('.cards').replaceChildren(...items);
    });
    restoreFocus(focus);
  }

  function focusedControl() {
    const active = document.activeElement;
    const item = active && active.closest && active.closest('[data-card-id]');
    if (!item) return null;
    return { id: item.dataset.cardId, action: active.dataset.action || active.id };
  }

  function restoreFocus(focus) {
    if (!focus) return;
    const item = ui.board.querySelector(`[data-card-id="${CSS.escape(focus.id)}"]`);
    if (!item) return;
    const target = item.querySelector(`[data-action="${CSS.escape(focus.action)}"]:enabled, #${CSS.escape(focus.action)}`) ||
      item.querySelector('button:enabled, input');
    if (target) target.focus();
  }

  // --- actions -------------------------------------------------------------

  function move(card, step) {
    const index = COLUMNS.indexOf(card.column);
    const to = COLUMNS[index + step];
    if (index < 0 || to === undefined) return; // edge control: never sent (ADR-0008)
    act(`Cannot move “${card.title}”`, 'POST', cardPath(card.id, '/move'), { to });
  }

  function remove(card) {
    act(`Cannot delete “${card.title}”`, 'DELETE', cardPath(card.id));
  }

  function startEdit(card) {
    state.editing = { id: card.id, title: card.title, description: card.description, invalid: false };
    render(state.board);
    const title = $('#edit-title');
    if (title) title.focus();
  }

  function cancelEdit() {
    const id = state.editing && state.editing.id;
    state.editing = null;
    render(state.board);
    const button = id && ui.board.querySelector(`[data-card-id="${CSS.escape(id)}"] [data-action="edit"]`);
    if (button) button.focus();
  }

  function saveEdit(card, title, description) {
    if (state.busy) return;
    // Keep the draft so a refused save re-opens the form with the user's text.
    state.editing = { id: card.id, title, description, invalid: false };
    act(`Cannot save “${card.title}”`, 'PATCH', cardPath(card.id), { title, description }, (outcome) => {
      if (outcome.ok) state.editing = null;
      else if (state.editing) state.editing.invalid = outcome.code === 'validation_failed';
    });
  }

  ui.board.addEventListener('click', (e) => {
    const button = e.target.closest('button[data-action]');
    if (!button || button.disabled || state.busy) return;
    const item = button.closest('[data-card-id]');
    const card = item && findCard(item.dataset.cardId);
    if (!card) return;
    switch (button.dataset.action) {
      case 'left': move(card, -1); break;
      case 'right': move(card, +1); break;
      case 'edit': startEdit(card); break;
      case 'delete': remove(card); break;
      case 'cancel': cancelEdit(); break;
    }
  });

  ui.addForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const payload = { title: ui.addTitle.value, description: ui.addDescription.value };
    act('Cannot add the card', 'POST', '/api/cards', payload, (outcome) => {
      if (outcome.ok) {
        ui.addForm.reset();
        ui.addTitle.removeAttribute('aria-invalid');
      } else if (outcome.code === 'validation_failed') {
        ui.addTitle.setAttribute('aria-invalid', 'true');
      }
      ui.addTitle.focus();
    });
  });

  // First load: the same render path as every action.
  (async () => {
    setBusy(true);
    try {
      await refresh('Cannot load the board');
    } finally {
      setBusy(false);
    }
  })();
})();
