import type { TRewardCategory } from '@/entities/reward';
import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_PROFILE } from './api-routes.ts';

export type TLevelRewardStatus = 'CLAIMED' | 'FROZEN' | 'UNOPENED' | 'LOCKED';

export type TLevelReward = {
  id: string;
  type: TRewardCategory;
  description: string;
};

export type TLevelRewardItem = {
  level: number;
  status: TLevelRewardStatus;
  reward: TLevelReward;
  expiresAt: string | null;
};

export const getLevelsRewards = async (
  signal?: AbortSignal,
): Promise<{ levels: TLevelRewardItem[] }> => {
  return await apiRequest(
    fetch(`${API_URL}${API_ROUTE_PROFILE.levels}`, {
      headers: getAuthHeaders(),
      signal,
    }),
    'Ошибка запроса getPetName',
  );
};

export const receiveLevelReward = async (
  rewardId: string,
  signal?: AbortSignal,
): Promise<{ levels: Pick<TLevelRewardItem, 'level' | 'status'> }> => {
  return await apiRequest(
    fetch(`${API_URL}${API_ROUTE_PROFILE.receiveLevelReward(rewardId)}`, {
      method: 'POST',
      headers: getAuthHeaders(),
      signal,
    }),
    'Ошибка запроса receiveLevelReward',
  );
};
