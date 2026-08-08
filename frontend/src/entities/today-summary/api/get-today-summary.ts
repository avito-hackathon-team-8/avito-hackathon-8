import type { TTaskType } from '@/entities/daily-task/api/tasks';
import type { TRewardCategory } from '@/entities/reward/api/rewards';

import { getTodaySummaryStats, type TTodaySummaryStats } from '../model/get-today-summary-stats ';
import { mockTodaySummary } from '../model/mock/mock-today-summary';

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
  rewardClaimed: boolean;
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

export const getTodaySummary = async (): Promise<TTodaySummaryStats> => {
  const data = await Promise.resolve(mockTodaySummary);

  return getTodaySummaryStats(data);
};
