import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { dailyTasksQueryKeys } from '../api/daily-tasks-keys';
import { getTasks, receiveTaskReward, type TTask } from '../api/tasks';

export type TReceiveTaskRewardVariables = {
  taskId: TTask['taskId'];
};

export const useDailyTasks = () => {
  const queryClient = useQueryClient();

  const queryKey = dailyTasksQueryKeys.list();

  const tasksQuery = useQuery({
    queryKey,
    queryFn: getTasks,
    retry: 1,
  });

  const receiveRewardMutation = useMutation({
    mutationFn: ({ taskId }: TReceiveTaskRewardVariables) => receiveTaskReward(taskId),

    onMutate: async ({ taskId }) => {
      await queryClient.cancelQueries({
        queryKey,
      });

      const previousTasks = queryClient.getQueryData<TTask[]>(queryKey);

      queryClient.setQueryData<TTask[]>(queryKey, (old) => {
        if (!old) {
          return old;
        }

        return old.map((task) =>
          task.taskId === taskId
            ? {
                ...task,
                status: 'CLAIMED' as const,
              }
            : task,
        );
      });

      return {
        previousTasks,
      };
    },

    onError: (_error, _variables, context) => {
      if (context?.previousTasks) {
        queryClient.setQueryData(queryKey, context.previousTasks);
      }

      toast.error('Произошла ошибка при получении награды за задание');
    },

    onSettled: () => {
      return queryClient.invalidateQueries({
        queryKey,
      });
    },
  });

  return {
    ...tasksQuery,

    receiveReward: receiveRewardMutation.mutate,
    isRewardPending: receiveRewardMutation.isPending,
  };
};
