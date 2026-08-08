import { useQuery } from '@tanstack/react-query';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getPetName } from '../api/get-pet';

export const usePetProfile = () => {
  return useQuery({
    queryKey: gamificationProfileKeys.pet(),
    queryFn: getPetName,
    retry: false,
    staleTime: 30_000,
  });
};
