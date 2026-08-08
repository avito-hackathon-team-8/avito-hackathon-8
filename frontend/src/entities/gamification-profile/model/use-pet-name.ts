import { useQuery } from '@tanstack/react-query';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getPetName } from '../api/pet';

export const usePetName = () => {
  return useQuery({
    queryKey: gamificationProfileKeys.pet(),
    queryFn: getPetName,
    retry: false,
    staleTime: 30_000,

    select: (pet) => pet.name,
  });
};
