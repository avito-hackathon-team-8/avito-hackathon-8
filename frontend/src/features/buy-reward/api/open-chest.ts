import type { TReward } from '@/entities/reward/api/rewards';
import { apiRequest } from '@/shared/api/api-request';
import { getAuthHeaders } from '@/shared/api/get-auth-headers';
import { API_URL } from '@/shared/config/api';

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
