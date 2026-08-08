import type { TRewardCategory } from '@/entities/reward/api/get-rewards.ts';
import { getAuthHeaders } from '@/shared/api/get-auth-headers.tsx';
import { API_URL } from '@/shared/config/api.ts';

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
  const response = await fetch(`${API_URL}${API_ROUTE_PROFILE.levels}`, {
    headers: getAuthHeaders(),
    signal,
  });

  if (!response.ok) {
    throw new Error(`Ошибка запроса getPetName: ${response.status}`);
  }

  return await response.json();
};
export const receiveLevelReward = async (
  rewardId: string,
  signal?: AbortSignal,
): Promise<{ levels: Pick<TLevelRewardItem, 'level' | 'status'> }> => {
  const response = await fetch(`${API_URL}${API_ROUTE_PROFILE.receiveLevelReward(rewardId)}`, {
    method: 'POST',
    headers: getAuthHeaders(),
    signal,
  });

  if (!response.ok) {
    throw new Error(`Ошибка запроса receiveLevelReward: ${response.status}`);
  }

  return await response.json();
};
