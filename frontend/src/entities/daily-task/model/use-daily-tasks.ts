import { useQuery } from '@tanstack/react-query';

import { tasksQueryKeys } from '../api/daily-tasks-keys';
import { getTasks } from '../api/get-tasks';

export const useDailyTasks = () => {
  return useQuery({
    queryKey: tasksQueryKeys.list(),
    queryFn: getTasks,
    retry: 2,
  });
};
