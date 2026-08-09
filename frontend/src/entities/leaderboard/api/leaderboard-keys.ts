import { mainQueryKey } from '@/shared/config';

export const leaderboardQueryKeys = {
  all: [mainQueryKey.all, 'leaderboard'] as const,
};
