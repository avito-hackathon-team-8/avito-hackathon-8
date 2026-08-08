import { mainQueryKey } from '@/shared/config/api';

export const leaderboardQueryKeys = {
  all: [mainQueryKey.all, 'leaderboard'] as const,
};
