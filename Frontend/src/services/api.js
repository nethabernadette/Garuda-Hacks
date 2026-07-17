const API_PORT = import.meta.env.VITE_API_PORT || '8080';
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || `http://localhost:${API_PORT}`;
const TOKEN_KEY = 'jalin_access_token';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

async function request(path, options = {}) {
  const token = getToken();
  const headers = new Headers(options.headers || {});
  const hasBody = options.body !== undefined && options.body !== null;

  if (hasBody && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  let response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
      body: hasBody && !(options.body instanceof FormData) ? JSON.stringify(options.body) : options.body,
    });
  } catch (error) {
    throw new ApiError('Backend tidak aktif atau jaringan bermasalah.', 0, error);
  }

  const contentType = response.headers.get('content-type') || '';
  const isHTML = contentType.includes('text/html');
  const payload = isHTML ? await response.text() : await safeJSON(response);

  if (!response.ok) {
    const message = payload?.error || payload?.message || statusMessage(response.status);
    throw new ApiError(message, response.status, payload);
  }

  return isHTML ? payload : payload?.data ?? payload;
}

async function safeJSON(response) {
  const text = await response.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch (error) {
    throw new ApiError('Response backend bukan JSON yang valid.', response.status, error);
  }
}

function statusMessage(status) {
  if (status === 401) return 'Sesi berakhir. Silakan masuk lagi.';
  if (status === 403) return 'Kamu tidak memiliki akses ke data ini.';
  if (status === 404) return 'Data tidak ditemukan.';
  if (status === 409) return 'Aksi belum dapat dilakukan karena status data belum sesuai.';
  if (status >= 500) return 'Server sedang bermasalah.';
  return 'Request gagal diproses.';
}

export class ApiError extends Error {
  constructor(message, status, details) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

export const api = {
  baseURL: API_BASE_URL,
  login: (body) => request('/auth/login', { method: 'POST', body }),
  register: (body) => request('/auth/register', { method: 'POST', body }),
  profile: () => request('/profile'),
  updateProfile: (body) => request('/profile', { method: 'PUT', body }),
  submitVerification: (body) => request('/profile/verification', { method: 'POST', body }),
  feed: (params = {}) => request(`/posts${query(params)}`),
  searchPosts: (params = {}) => request(`/posts/search${query(params)}`),
  createSupply: (body) => request('/posts/supply', { method: 'POST', body }),
  createDemand: (body) => request('/posts/demand', { method: 'POST', body }),
  notifications: (params = {}) => request(`/notifications${query(params)}`),
  unreadCount: () => request('/notifications/unread-count'),
  markNotificationRead: (id) => request(`/notifications/${id}/read`, { method: 'PATCH' }),
  markAllNotificationsRead: () => request('/notifications/read-all', { method: 'PATCH' }),
  matches: (params = {}) => request(`/matches${query(params)}`),
  createInterestMatch: (body) => request('/matches/interest', { method: 'POST', body }),
  agreements: () => request('/agreements'),
  createAgreement: (body) => request('/agreements', { method: 'POST', body }),
  confirmAgreement: (id) => request(`/agreements/${id}/confirm`, { method: 'POST', body: {} }),
  demoConfirmAgreement: (id) => request(`/agreements/${id}/demo-confirm`, { method: 'POST', body: {} }),
  agreementDocument: (id) => request(`/agreements/${id}/document`),
  agreementDocumentHTML: (id) => request(`/agreements/${id}/document/html`),
  agreementContact: (id) => request(`/agreements/${id}/contact`),
};

function query(params) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, value);
    }
  });
  const value = search.toString();
  return value ? `?${value}` : '';
}
