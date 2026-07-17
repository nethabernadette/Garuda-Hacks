import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { api, clearToken, getToken, setToken } from '../services/api.js';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setTokenState] = useState(getToken());
  const [profile, setProfile] = useState(null);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  async function refreshProfile() {
    if (!getToken()) return null;
    setStatus('loading');
    try {
      const data = await api.profile();
      setProfile(data);
      setError('');
      setStatus('success');
      return data;
    } catch (err) {
      setError(err.message);
      setStatus('error');
      return null;
    }
  }

  async function login(email, password) {
    setStatus('loading');
    const data = await api.login({ email, password });
    setToken(data?.access_token);
    setTokenState(data?.access_token);
    setStatus('success');
    await refreshProfile();
    return data;
  }

  async function register(role, email, password) {
    setStatus('loading');
    await api.register({ role, email, password });
    return login(email, password);
  }

  function logout() {
    clearToken();
    setTokenState(null);
    setProfile(null);
  }

  useEffect(() => {
    if (token) refreshProfile();
  }, [token]);

  const value = useMemo(
    () => ({ token, profile, status, error, login, register, logout, refreshProfile, setProfile }),
    [token, profile, status, error],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}
