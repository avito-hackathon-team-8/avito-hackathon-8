import { useCallback, useEffect, useRef, useState } from 'react';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { gamificationProfileKeys, type TPet, usePetProfile } from '@/entities/gamification-profile';
import { rewardsQueryKeys, type TReward } from '@/entities/reward';

import { addMVPLeaves } from '../api/add-mvp-leaves';
import { openChest as openChestRequest } from '../api/open-chest';

export const useBuyReward = () => {
  const queryClient = useQueryClient();
  const { data: pet } = usePetProfile();
  const [reward, setReward] = useState<TReward | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [isRewardVisible, setIsRewardVisible] = useState(false);
  const rewardTimerRef = useRef<number | null>(null);

  const clearRewardTimer = useCallback(() => {
    if (rewardTimerRef.current !== null) {
      clearTimeout(rewardTimerRef.current);
      rewardTimerRef.current = null;
    }
  }, []);

  useEffect(() => clearRewardTimer, [clearRewardTimer]);

  const { mutate, isPending } = useMutation({
    mutationFn: openChestRequest,
    onSuccess: (receivedReward) => {
      clearRewardTimer();
      setReward(receivedReward);
      setIsRewardVisible(false);
      setIsOpen(true);

      rewardTimerRef.current = setTimeout(() => {
        setIsRewardVisible(true);
        rewardTimerRef.current = null;
      }, 1_000);

      void queryClient.invalidateQueries({
        queryKey: rewardsQueryKeys.list(),
      });
    },
    onError: () => {
      toast.error('Не удалось открыть сундук');
    },
  });

  const { mutate: addMVPLeavesRequest, isPending: isMVPLeavesPending } = useMutation({
    mutationFn: addMVPLeaves,
    onSuccess: (updatedPet) => {
      queryClient.setQueryData<TPet>(gamificationProfileKeys.pet(), (currentPet) =>
        currentPet ? { ...currentPet, ...updatedPet } : currentPet,
      );
    },
    onError: () => {
      toast.error('Не удалось начислить листья');
    },
  });

  const openChest = useCallback(() => {
    mutate();
  }, [mutate]);

  const closeModal = useCallback(() => {
    clearRewardTimer();
    setIsOpen(false);
    setIsRewardVisible(false);
    setReward(null);
  }, [clearRewardTimer]);

  const isDisabled =
    !pet || pet.nextLevelTargetLeaves !== 0 || pet.leaves < pet.chestPrice || isPending;

  return {
    pet,
    reward,
    isOpen,
    isPending,
    isDisabled,
    isMVPLeavesPending,
    isRewardVisible,
    openChest,
    addMVPLeaves: () => addMVPLeavesRequest(),
    closeModal,
  };
};
