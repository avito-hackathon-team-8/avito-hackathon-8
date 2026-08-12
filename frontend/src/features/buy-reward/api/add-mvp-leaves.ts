import type { TPet } from '@/entities/gamification-profile';
import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

export type TMVPLeavesResponse = Omit<TPet, 'bedImageUrl' | 'bowlImageUrl'>;

export const addMVPLeaves = async (): Promise<TMVPLeavesResponse> => {
  return await apiRequest(
    fetch(`${API_URL}/v1/pet/mvp/leaves`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
    'Не удалось начислить листья',
  );
};
