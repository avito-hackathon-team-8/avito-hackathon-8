import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_PROFILE } from './rewards-routes';

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

type TResponse = {
  groups: TRewardGroup[];
};

export const getRewards = async (): Promise<TResponse> => {
  return await apiRequest(
    fetch(`${API_URL}${API_ROUTE_PROFILE.rewards}`, {
      headers: getAuthHeaders(),
    }),
  );
};
