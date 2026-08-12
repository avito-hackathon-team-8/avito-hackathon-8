import { useQuery } from '@tanstack/react-query';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getPet } from '../api/pet';

export const usePetName = () => {
  return useQuery({
    queryKey: gamificationProfileKeys.pet(),
    queryFn: getPet,
    retry: false,
    staleTime: 30_000,

    select: (pet) => pet.name,
  });
};
