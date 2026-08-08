import { useQuery } from '@tanstack/react-query';

import { getLeaderBoard } from '../api/leaderboard';
import { leaderboardQueryKeys } from '../api/leaderboard-keys';

export const useLeaderboard = () => {
  return useQuery({
    queryKey: leaderboardQueryKeys.all,
    queryFn: getLeaderBoard,
    retry: 1,
  });
};
