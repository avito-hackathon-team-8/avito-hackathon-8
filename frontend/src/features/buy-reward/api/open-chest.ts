import type { TReward } from '@/entities/reward';
import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

const OPEN_CHEST_ROUTE = '/v1/pet/chests/open';

export const openChest = async (): Promise<TReward> => {
  return await apiRequest(
    fetch(`${API_URL}${OPEN_CHEST_ROUTE}`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
    'Не удалось открыть сундук',
  );
};
