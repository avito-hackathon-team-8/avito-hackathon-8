import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_LEADERBOARD } from './api-route';

export type TLeaderboardUser = {
  playerId: string;
  nickname: string;
  position: number;
  leaves: number;
};

export type TResponseLeaderboard = {
  period: {
    key: string;
    startAt: string;
    endAt: string;
  };
  calculatedAt: string;
  nextCalculationAt: string;
  items: TLeaderboardUser[];
};

export const getLeaderBoard = async (): Promise<TResponseLeaderboard> => {
  return await apiRequest(
    fetch(`${API_URL}${API_ROUTE_LEADERBOARD.list}`, {
      headers: getAuthHeaders(),
    }),
  );
};
