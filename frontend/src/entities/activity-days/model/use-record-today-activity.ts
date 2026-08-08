import { useEffect, useRef } from 'react';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { recordTodayActivity, type TResponseActivityDay } from '../api/activity-day';
import { activityDayKeys } from '../api/activity-days-keys';

export const useRecordTodayActivity = () => {
  const queryClient = useQueryClient();

  const queryKey = activityDayKeys.week();

  const wasRecorded = useRef(false);

  const recordTodayActivityMutation = useMutation({
    mutationFn: recordTodayActivity,

    onMutate: async () => {
      await queryClient.cancelQueries({
        queryKey,
      });

      const previousActivityDays = queryClient.getQueryData<TResponseActivityDay>(queryKey);
      const today = new Date().toISOString().slice(0, 10);

      queryClient.setQueryData<TResponseActivityDay>(queryKey, (old) => {
        if (!old) {
          return old;
        }

        return {
          ...old,
          claims: old.claims.map((day) =>
            day.date === today && day.status === 'FUTURE'
              ? {
                  ...day,
                  status: 'AVAILABLE' as const,
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

      toast.error('Не удалось отметить активность за сегодня');
    },

    onSuccess: () => {
      return queryClient.invalidateQueries({
        queryKey,
      });
    },
  });

  useEffect(() => {
    if (wasRecorded.current) {
      return;
    }

    wasRecorded.current = true;
    recordTodayActivityMutation.mutate();
  }, [recordTodayActivityMutation]);
};
