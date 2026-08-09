import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { rewardsQueryKeys } from '@/entities/reward';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import { getLevelsRewards, receiveLevelReward, type TLevelRewardItem } from '../api/levels-rewards';

type TLevelsData = {
  levels: TLevelRewardItem[];
};

export type TReceiveRewardVariables = {
  rewardId: TLevelRewardItem['reward']['id'];
  level: TLevelRewardItem['level'];
};

export const useLevelsProfile = () => {
  const queryClient = useQueryClient();

  const queryKey = gamificationProfileKeys.levels();

  const levelsQuery = useQuery({
    queryKey,
    queryFn: ({ signal }) => getLevelsRewards(signal),
    retry: 1,
  });

  const receiveRewardMutation = useMutation({
    mutationFn: ({ rewardId }: TReceiveRewardVariables) => receiveLevelReward(rewardId),

    onMutate: async ({ level }) => {
      await queryClient.cancelQueries({
        queryKey,
      });

      const previousLevels = queryClient.getQueryData<TLevelsData>(queryKey);

      queryClient.setQueryData<TLevelsData>(queryKey, (old) => {
        if (!old) return old;

        return {
          ...old,
          levels: old.levels.map((item) =>
            item.level === level
              ? {
                  ...item,
                  status: 'CLAIMED' as const,
                }
              : item,
          ),
        };
      });

      return {
        previousLevels,
      };
    },

    onError: (_error, _variables, context) => {
      if (context?.previousLevels) {
        queryClient.setQueryData(queryKey, context.previousLevels);
      }

      toast.error('Произошла ошибка при получении награды за уровень');
    },

    onSuccess: () => {
      return queryClient.invalidateQueries({
        queryKey: rewardsQueryKeys.list(),
      });
    },

    onSettled: () => {
      return queryClient.invalidateQueries({
        queryKey,
      });
    },
  });

  return {
    ...levelsQuery,

    receiveReward: receiveRewardMutation.mutate,
    isRewardPending: receiveRewardMutation.isPending,
  };
};
