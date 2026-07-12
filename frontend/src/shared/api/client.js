import { API_BASE_URL } from "../../app/config.js";

export function getJSON(path, fallback = null, token) {
  return request(path, { token, fallback });
}

export function postJSON(path, payload, token) {
  return request(path, { method: "POST", payload, token });
}

export function putJSON(path, payload, token) {
  return request(path, { method: "PUT", payload, token });
}

export function deleteJSON(path, payload, token) {
  return request(path, { method: "DELETE", payload, token });
}

export async function getText(path, token) {
  try {
    const response = await fetch(apiURL(path), { headers: authHeaders(token) });
    if (!response.ok) {
      return { ok: false, data: "", error: await responseError(response), status: response.status };
    }
    return { ok: true, data: await response.text(), error: "", status: response.status };
  } catch {
    return { ok: false, data: "", error: "API inaccessible.", status: 0 };
  }
}

async function request(path, options = {}) {
  const { method = "GET", payload, token, fallback = null } = options;
  try {
    const response = await fetch(apiURL(path), {
      method,
      headers: {
        ...(payload !== undefined ? { "Content-Type": "application/json" } : {}),
        ...authHeaders(token)
      },
      body: payload !== undefined ? JSON.stringify(payload) : undefined
    });

    if (!response.ok) {
      return {
        ok: false,
        fromAPI: true,
        data: fallback,
        error: await responseError(response),
        status: response.status
      };
    }

    const data = response.status === 204 ? null : await response.json();
    return { ok: true, fromAPI: true, data, error: "", status: response.status };
  } catch {
    return { ok: false, fromAPI: false, data: fallback, error: "API inaccessible.", status: 0 };
  }
}

async function responseError(response) {
  try {
    const payload = await response.json();
    return payload.error || `Erreur HTTP ${response.status}`;
  } catch {
    return `Erreur HTTP ${response.status}`;
  }
}

function authHeaders(token) {
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function apiURL(path) {
  const rootPath = ["/user", "/admin", "/study", "/search", "/openapi", "/swagger"]
    .some((prefix) => path.startsWith(prefix));
  const base = rootPath ? API_BASE_URL.replace(/\/api$/, "") : API_BASE_URL;
  return `${base}${path}`;
}
