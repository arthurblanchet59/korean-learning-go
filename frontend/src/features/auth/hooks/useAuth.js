import { useCallback, useEffect, useState } from "react";

import { fetchCurrentUser, loginUser, registerUser, updateCurrentUser } from "../../../shared/api/authApi.js";

const tokenStorageKey = "korean-learning-token";

export function useAuth() {
  const [token, setToken] = useState(() => localStorage.getItem(tokenStorageKey));
  const [user, setUser] = useState(null);
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(Boolean(token));

  const persistSession = useCallback((result) => {
    localStorage.setItem(tokenStorageKey, result.token);
    setToken(result.token);
    setUser(result.user);
    setError("");
  }, []);

  const login = useCallback(
    async (payload) => {
      setIsLoading(true);
      const result = await loginUser(payload);
      setIsLoading(false);

      if (!result.ok) {
        setError("Identifiants invalides.");
        return false;
      }

      persistSession(result.data);
      return true;
    },
    [persistSession]
  );

  const register = useCallback(
    async (payload) => {
      setIsLoading(true);
      const result = await registerUser(payload);
      setIsLoading(false);

      if (!result.ok) {
        setError("Impossible de creer ce compte.");
        return false;
      }

      persistSession(result.data);
      return true;
    },
    [persistSession]
  );

  const logout = useCallback(() => {
    localStorage.removeItem(tokenStorageKey);
    setToken(null);
    setUser(null);
    setError("");
  }, []);

  const updateProfile = useCallback(
    async (payload) => {
      const result = await updateCurrentUser(payload, token);
      if (!result.ok) {
        setError("Impossible de mettre a jour le profil.");
        return false;
      }

      setUser(result.data);
      setError("");
      return true;
    },
    [token]
  );

  useEffect(() => {
	window.addEventListener("auth:unauthorized", logout);
	return () => window.removeEventListener("auth:unauthorized", logout);
  }, [logout]);

  useEffect(() => {
    if (!token) {
      setIsLoading(false);
      return;
    }

    let active = true;
    setIsLoading(true);
    fetchCurrentUser(token).then((result) => {
      if (!active) {
        return;
      }

      setIsLoading(false);
      if (!result.fromAPI || !result.data) {
        logout();
        return;
      }

      setUser(result.data);
    });

    return () => {
      active = false;
    };
  }, [logout, token]);

  return {
    token,
    user,
    error,
    isLoading,
    login,
    logout,
    register,
    updateProfile
  };
}
