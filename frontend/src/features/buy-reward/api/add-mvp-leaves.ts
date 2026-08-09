import type { TPet } from '@/entities/gamification-profile';
import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

export const addMVPLeaves = async (): Promise<TPet> => {
  return await apiRequest(
    fetch(`${API_URL}/v1/pet/mvp/leaves`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
    'Не удалось начислить листья',
  );
};
