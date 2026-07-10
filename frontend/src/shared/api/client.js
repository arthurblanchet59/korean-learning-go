import { API_BASE_URL } from "../../app/config.js";

export async function getJSON(path, fallback, token) {
  try {
    const response = await fetch(apiURL(path), {
      headers: authHeaders(token)
    });
    if (!response.ok) {
      return { data: fallback, fromAPI: false };
    }

    return { data: await response.json(), fromAPI: true };
  } catch {
    return { data: fallback, fromAPI: false };
  }
}

export async function postJSON(path, payload, token) {
  try {
    const response = await fetch(apiURL(path), {
      method: "POST",
      headers: { "Content-Type": "application/json", ...authHeaders(token) },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      return { ok: false, data: null };
    }

    return { ok: true, data: await response.json() };
  } catch {
    return { ok: false, data: null };
  }
}

export async function putJSON(path, payload, token) {
  try {
    const response = await fetch(apiURL(path), {
      method: "PUT",
      headers: { "Content-Type": "application/json", ...authHeaders(token) },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      return { ok: false, data: null };
    }

    return { ok: true, data: await response.json() };
  } catch {
    return { ok: false, data: null };
  }
}

function authHeaders(token) {
  if (!token) {
    return {};
  }

  return { Authorization: `Bearer ${token}` };
}

function apiURL(path) {
  if (path.startsWith("/user") || path.startsWith("/admin")) {
    return `${API_BASE_URL.replace(/\/api$/, "")}${path}`;
  }

  return `${API_BASE_URL}${path}`;
}
