import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { getActivityDay, receiveActivityDayReward, type TActivityDay } from '../api/activity-day';
import { activityDayKeys } from '../api/activity-days-keys';

export type TReceiveActivityDayRewardVariables = {
  claimId: NonNullable<TActivityDay['claimId']>;
};

type TUseActivityDaysOptions = {
  enabled?: boolean;
};

export const useActivityDays = ({ enabled = true }: TUseActivityDaysOptions = {}) => {
  const queryClient = useQueryClient();

  const queryKey = activityDayKeys.week();

  const activityDaysQuery = useQuery({
    queryKey,
    queryFn: getActivityDay,
    enabled,
    retry: 1,
  });

  const receiveRewardMutation = useMutation({
    mutationFn: receiveActivityDayReward,

    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey,
        refetchType: 'none',
      });

      try {
        await queryClient.fetchQuery({
          queryKey,
          queryFn: getActivityDay,
        });
      } catch {
        toast.error('Не удалось обновить данные за неделю');
      }
    },

    onError: () => {
      toast.error('Произошла ошибка при получении ежедневной награды');
    },
  });

  return {
    ...activityDaysQuery,

    receiveReward: receiveRewardMutation.mutate,
    isRewardPending: receiveRewardMutation.isPending,
  };
};
