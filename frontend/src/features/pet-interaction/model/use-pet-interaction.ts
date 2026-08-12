import { useCallback } from 'react';

import { useMutation } from '@tanstack/react-query';

import { usePetProfile } from '@/entities/gamification-profile';

import { carePetPost, type TCarePetBody } from '../api/pet-interactions';

export const usePetInteraction = () => {
  const { data: pet, updatePetProfile } = usePetProfile();

  const { mutate, isPending } = useMutation({
    mutationFn: (body: TCarePetBody) => carePetPost(body),
    onSuccess: (petState) => {
      if (!pet) {
        return;
      }

      updatePetProfile({
        ...pet,
        ...petState,
      });
    },
  });

  const carePet = useCallback(
    (type: TCarePetBody['type']) => {
      mutate({ type });
    },
    [mutate],
  );

  return {
    carePet,
    isPending,
  };
};
