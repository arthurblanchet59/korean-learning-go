import { getJSON, postJSON, putJSON } from "./client.js";

export function registerUser(payload) {
  return postJSON("/user/register", payload);
}

export function loginUser(payload) {
  return postJSON("/user/login", payload);
}

export function fetchCurrentUser(token) {
  return getJSON("/user/me", null, token);
}

export function updateCurrentUser(payload, token) {
  return putJSON("/user/me", payload, token);
}
