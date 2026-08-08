import { useQuery } from '@tanstack/react-query';

import { getRewards } from '../api/rewards';
import { rewardsQueryKeys } from '../api/rewards-keys';

export const useRewards = () => {
  const queryKey = rewardsQueryKeys.list();

  return useQuery({
    queryKey,
    queryFn: getRewards,
    retry: 1,
  });
};
