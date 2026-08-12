import { useCallback } from 'react';

import { useQuery, useQueryClient } from '@tanstack/react-query';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getPet, type TPet } from '../api/pet';

export const usePetProfile = () => {
  const queryClient = useQueryClient();

  const updatePetProfile = useCallback(
    (pet: TPet | null) => {
      queryClient.setQueryData(gamificationProfileKeys.pet(), pet);
    },
    [queryClient],
  );

  return {
    ...useQuery({
      queryKey: gamificationProfileKeys.pet(),
      queryFn: getPet,
      retry: false,
      staleTime: 30_000,
    }),
    updatePetProfile,
  };
};
