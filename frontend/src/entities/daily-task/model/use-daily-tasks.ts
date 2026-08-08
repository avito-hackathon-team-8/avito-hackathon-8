import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { dailyTasksQueryKeys } from '../api/daily-tasks-keys';
import { getTasks, receiveTaskReward, type TTask } from '../api/tasks';

type TTasksData = {
  tasks: TTask[];
};

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

      const previousTasks = queryClient.getQueryData<TTasksData>(queryKey);

      queryClient.setQueryData<TTasksData>(queryKey, (old) => {
        if (!old) {
          return old;
        }

        return {
          ...old,
          tasks: old.tasks.map((task) =>
            task.taskId === taskId
              ? {
                  ...task,
                  status: 'CLAIMED' as const,
                }
              : task,
          ),
        };
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
