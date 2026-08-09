import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_PROFILE } from './api-routes.ts';

export type TPet = {
  name: string;
  level: number;
  leaves: number;
  nextLevelTargetLeaves?: number;
  levelUp?: boolean;
  chestPrice: number;
};

export const getPetName = async (): Promise<TPet> => {
  const response = await fetch(`${API_URL}/${API_ROUTE_PROFILE.petName}`, {
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error(`Ошибка запроса getPetName: ${response.status}`);
  }

  return await response.json();
};

export const updatePetName = async (name: string): Promise<TPet> => {
  return await apiRequest(
    fetch(`${API_URL}/${API_ROUTE_PROFILE.petName}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
      },
      body: JSON.stringify({ name }),
    }),
    'Не удалось сохранить имя питомца',
  );
};
