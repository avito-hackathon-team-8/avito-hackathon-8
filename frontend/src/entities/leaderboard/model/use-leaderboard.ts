import { useQuery } from '@tanstack/react-query';

import { getLeaderBoard } from '../api/leaderboard';
import { leaderboardQueryKeys } from '../api/leaderboard-keys';

export const LEADERBOARD_REFRESH_INTERVAL = 10 * 60 * 1000;

export const useLeaderboard = () => {
  return useQuery({
    queryKey: leaderboardQueryKeys.all,
    queryFn: getLeaderBoard,
    retry: 1,
    refetchInterval: LEADERBOARD_REFRESH_INTERVAL,
    refetchIntervalInBackground: true,
  });
};
