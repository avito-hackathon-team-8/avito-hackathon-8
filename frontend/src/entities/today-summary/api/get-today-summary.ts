import type { TTaskType } from '@/entities/daily-task/api/tasks';
import type { TRewardCategory } from '@/entities/reward/api/rewards';
import { apiRequest } from '@/shared/api/api-request';
import { getAuthHeaders } from '@/shared/api/get-auth-headers';
import { API_URL } from '@/shared/config/api';

import { API_ROUTE_TODAY_SUMMARY } from './api-routes';

export type TTodaySummaryReward = {
  rewardId: string;
  type: TRewardCategory;
  title: string;
  expiresAt: string;
  receivedAt: string;
};

export type TTodaySummaryTask = {
  taskId: string;
  type: TTaskType;
  description: string;
  rewardLeaves: number;
  completedAt: string;
};

export type TTodaySummaryLevelUp = {
  fromLevel: number;
  toLevel: number;
  occurredAt: string;
};

export type TTodaySummary = {
  leavesEarnedToday: number;
  date: string;
  rewards: TTodaySummaryReward[];
  levelUp: TTodaySummaryLevelUp | null;
  tasks: TTodaySummaryTask[];
  visitedToday: boolean;
  updatedAt: string;
};

export const getTodaySummary = async (): Promise<TTodaySummary> => {
  return apiRequest(
    fetch(`${API_URL}${API_ROUTE_TODAY_SUMMARY.report}`, {
      headers: getAuthHeaders(),
    }),
  );
};
