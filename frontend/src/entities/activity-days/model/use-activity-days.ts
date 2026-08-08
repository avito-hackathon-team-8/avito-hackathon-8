import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import {
  getActivityDay,
  receiveActivityDayReward,
  type TActivityDay,
  type TResponseActivityDay,
} from '../api/activity-day';
import { activityDayKeys } from '../api/activity-days-keys';

export type TReceiveActivityDayRewardVariables = {
  claimId: NonNullable<TActivityDay['claimId']>;
};

export const useActivityDays = () => {
  const queryClient = useQueryClient();

  const queryKey = activityDayKeys.week();

  const activityDaysQuery = useQuery({
    queryKey,
    queryFn: getActivityDay,
    retry: 2,
  });

  const receiveRewardMutation = useMutation({
    mutationFn: () => receiveActivityDayReward(),

    onMutate: async ({ claimId }: TReceiveActivityDayRewardVariables) => {
      await queryClient.cancelQueries({
        queryKey,
      });

      const previousActivityDays = queryClient.getQueryData<TResponseActivityDay>(queryKey);

      queryClient.setQueryData<TResponseActivityDay>(queryKey, (old) => {
        if (!old) {
          return old;
        }

        return {
          ...old,

          claimedDaysCount: Math.min(
            old.claimedDaysCount + 1,
            7,
          ) as TResponseActivityDay['claimedDaysCount'],

          claims: old.claims.map((day) =>
            day.claimId === claimId
              ? {
                  ...day,
                  status: 'CLAIMED' as const,
                }
              : day,
          ),
        };
      });

      return {
        previousActivityDays,
      };
    },

    onError: (_error, _variables, context) => {
      if (context?.previousActivityDays) {
        queryClient.setQueryData(queryKey, context.previousActivityDays);
      }

      toast.error('Произошла ошибка при получении ежедневной награды');
    },

    onSettled: () => {
      return queryClient.invalidateQueries({
        queryKey,
      });
    },
  });

  return {
    ...activityDaysQuery,

    receiveReward: receiveRewardMutation.mutate,
    isRewardPending: receiveRewardMutation.isPending,
  };
};
