import { getSessionStorageValue, sessionStorageKeysMap } from '../lib/session-storage';

export const getAuthHeaders = (): HeadersInit => {
  const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

  return token ? { Authorization: `Bearer ${token}` } : {};
};
