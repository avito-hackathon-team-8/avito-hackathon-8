import { getSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib';

export const getAuthHeaders = (): HeadersInit => {
  const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

  return token ? { Authorization: `Bearer ${token}` } : {};
};
