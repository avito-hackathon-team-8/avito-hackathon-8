import { mockRewards } from '../model/mock/mockRewards';
export type TRewardCategory =
  'AVITO_BONUS' | 'FREE_DELIVERY' | 'FREE_PROMOTION' | 'PROMOTION_DISCOUNT' | 'DELIVERY_DISCOUNT';

export type TRewardSource = 'LEVEL' | 'CHEST' | 'LEADERBOARD';

export type TRewardStatus = 'ACTIVE' | 'REDEEMED' | 'EXPIRED';

export type TReward = {
  id: string;
  title: string;
  category: TRewardCategory;
  categoryName: string;
  source: TRewardSource;
  active: boolean;
  status: TRewardStatus;
  expiresAt: string;
  awardedAt: string;
  redeemedAt: string | null;
};

export type TRewardGroup = {
  category: TRewardCategory;
  categoryName: string;
  items: TReward[];
};

export const getTasks = async (): Promise<TRewardGroup[]> => {
  return await Promise.resolve(mockRewards);
};
