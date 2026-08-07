import { useQuery } from '@tanstack/react-query';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getLevelsRewards } from '../api/get-levels-reward';

export const useLevelsProfile = () => {
  return useQuery({
    queryKey: gamificationProfileKeys.levels(),
    queryFn: getLevelsRewards,
    retry: 2,
  });
};
