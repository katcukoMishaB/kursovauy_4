const API_BASE = '/api';

async function request(method, path, body, isBlob) {
  const headers = { 'Content-Type': 'application/json' };
  const token = localStorage.getItem('token');
  if (token) headers['Authorization'] = 'Bearer ' + token;

  const res = await fetch(API_BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (isBlob) {
    if (!res.ok) throw new Error('Ошибка запроса');
    return await res.blob();
  }

  let data = null;
  const text = await res.text();
  if (text) {
    try { data = JSON.parse(text); } catch { data = { raw: text }; }
  }

  if (!res.ok) {
    const msg = (data && data.error) || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

async function uploadFile(file) {
  const fd = new FormData();
  fd.append('file', file);
  const headers = {};
  const token = localStorage.getItem('token');
  if (token) headers['Authorization'] = 'Bearer ' + token;
  const res = await fetch(API_BASE + '/files/upload', { method: 'POST', headers, body: fd });
  const text = await res.text();
  if (!res.ok) {
    throw new Error('HTTP ' + res.status + (text ? ': ' + text.slice(0, 200) : ''));
  }
  try {
    return JSON.parse(text);
  } catch {
    throw new Error('Сервер вернул не JSON: ' + text.slice(0, 200));
  }
}

window.api = {
  get: (p) => request('GET', p),
  post: (p, b) => request('POST', p, b),
  put: (p, b) => request('PUT', p, b),
  del: (p) => request('DELETE', p),
  blob: (p) => request('GET', p, null, true),
  upload: uploadFile,
  base: API_BASE,
};

window.fileLabel = (value) => {
  if (!value) return '';
  if (value.startsWith('/api/files/')) {
    const tail = value.substring('/api/files/'.length);
    const sep = tail.indexOf('__');
    return sep >= 0 ? tail.substring(sep + 2) : tail;
  }
  return value;
};

window.isValidUrl = (v) => {
  if (!v) return false;
  const s = String(v).trim();
  return /^https?:\/\/.+\..+/i.test(s) || s.startsWith('/api/files/');
};

window.formatDate = (d) => d ? new Date(d).toLocaleDateString('ru-RU') : '';
window.statusColor = (s) => {
  if (s === 'активен' || s === 'в работе' || s === 'одобрена') return 'bg-emerald-100 text-emerald-700';
  if (s === 'завершён' || s === 'завершена') return 'bg-indigo-100 text-indigo-700';
  if (s === 'архивирован' || s === 'отклонена') return 'bg-slate-200 text-slate-600';
  if (s === 'новая' || s === 'в рассмотрении') return 'bg-amber-100 text-amber-700';
  return 'bg-slate-100 text-slate-600';
};
window.roleLabel = (r) => ({admin:'Администратор',organizer:'Организатор',participant:'Участник'}[r] || r);

window.paginate = (list, page, pageSize) => {
  const ps = pageSize || 50;
  const p = Math.max(1, page || 1);
  return (list || []).slice((p - 1) * ps, p * ps);
};
