import { useEffect, useRef } from 'react';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { getActivityDay, recordTodayActivity } from '../api/activity-day';
import { activityDayKeys } from '../api/activity-days-keys';

export const useRecordTodayActivity = () => {
  const queryClient = useQueryClient();
  const wasRecorded = useRef(false);

  const queryKey = activityDayKeys.week();

  const mutation = useMutation({
    mutationFn: recordTodayActivity,

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
        toast.error('Не удалось получить данные за неделю');
      }
    },

    onError: () => {
      toast.error('Не удалось отметить активность за сегодня');
    },
  });

  const { mutate } = mutation;

  useEffect(() => {
    if (wasRecorded.current) {
      return;
    }

    wasRecorded.current = true;
    mutate();
  }, [mutate]);

  return mutation;
};
