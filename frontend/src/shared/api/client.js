import { API_BASE_URL } from "../../app/config.js";

export async function getJSON(path, fallback) {
  try {
    const response = await fetch(`${API_BASE_URL}${path}`);
    if (!response.ok) {
      return { data: fallback, fromAPI: false };
    }

    return { data: await response.json(), fromAPI: true };
  } catch {
    return { data: fallback, fromAPI: false };
  }
}

export async function postJSON(path, payload) {
  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
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
