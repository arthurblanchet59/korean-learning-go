import { afterEach, describe, expect, it, vi } from "vitest";

import { getJSON } from "./client.js";

describe("API client", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("notifies the application when an authenticated request expires", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: "expired" }), {
      status: 401,
      headers: { "Content-Type": "application/json" }
    })));
    const listener = vi.fn();
    window.addEventListener("auth:unauthorized", listener, { once: true });

    const result = await getJSON("/cards", [], "expired-token");

    expect(result.status).toBe(401);
    expect(listener).toHaveBeenCalledOnce();
  });
});

