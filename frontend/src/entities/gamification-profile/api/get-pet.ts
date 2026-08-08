import { API_URL } from '@/shared/config/api.ts';
import { getSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib/session-storage';

import { API_ROUTE_PROFILE } from '../model/api-routes.ts';

export type TPet = {
  name: string;
  level: number;
  leaves: number;
};

const getAuthHeaders = (): HeadersInit => {
  const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

  return token ? { Authorization: `Bearer ${token}` } : {};
};

export const getPetName = async (): Promise<TPet> => {
  const response = await fetch(`${API_URL}${API_ROUTE_PROFILE.petName}`, {
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Ошибка запроса getPetName: ${response.status}`);
  }

  return response.json();
};

export const updatePetName = async (name: string): Promise<TPet> => {
  const response = await fetch(`${API_URL}${API_ROUTE_PROFILE.petName}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || 'Не удалось сохранить имя питомца');
  }

  return response.json();
};
