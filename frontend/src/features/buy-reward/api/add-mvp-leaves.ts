import type { TPet } from '@/entities/gamification-profile/api/pet';
import { apiRequest } from '@/shared/api/api-request';
import { getAuthHeaders } from '@/shared/api/get-auth-headers';
import { API_URL } from '@/shared/config/api';

export const addMVPLeaves = async (): Promise<TPet> => {
  return await apiRequest(
    fetch(`${API_URL}/v1/pet/mvp/leaves`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
    'Не удалось начислить листья',
  );
};
