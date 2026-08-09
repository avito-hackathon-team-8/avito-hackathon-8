import { apiRequest } from '@/shared/api';
import { removeSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib';

const CURRENT_USER_API = '/api/app/auth/me';

export type User = {
  id: string;
  email: string;
  verified: boolean;
  leaderboard?: {
    period: {
      key: string;
      startAt: string;
      endAt: string;
    };
    calculatedAt: string;
    nextCalculationAt: string;
    player: {
      playerId: string;
      nickname: string;
      position: number;
      leaves: number;
      isTop10: boolean;
    };
  };
};

export const getCurrentUser = async (token: string | null): Promise<User> => {
  return await apiRequest(
    fetch(CURRENT_USER_API, {
      headers: { Authorization: `Bearer ${token}` },
    }),
    'Ошибка получения данных getCurrentUser',
    () => removeSessionStorageValue(sessionStorageKeysMap.authToken),
  );
};
