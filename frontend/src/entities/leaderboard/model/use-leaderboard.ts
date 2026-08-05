import { useQuery } from "@tanstack/react-query";

import { getLeaderBoard } from "../api/get-leaderboard";
import { leaderboardKeys } from "../api/leaderboard-keys";

export const useLeaderboard = () => {
  return useQuery({
    queryKey: leaderboardKeys.all,
    queryFn: getLeaderBoard,
    retry: 2,
  });
};
