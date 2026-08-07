import type { TRewardCategory } from '@/entities/reward/api/get-rewards.ts';

import { mockLevelRewards } from '../model/mock/mock-level-rewards.ts';

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

export const getLevelsRewards = async (): Promise<TLevelRewardItem[]> => {
  return await Promise.resolve(mockLevelRewards);
};
