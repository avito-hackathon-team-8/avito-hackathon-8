import { useQuery } from '@tanstack/react-query';

import { getTodaySummary } from '../api/get-today-summary';

export const todaySummaryQueryKeys = {
  all: ['today-summary'] as const,
  current: () => [...todaySummaryQueryKeys.all, 'current'] as const,
};

export const useTodaySummary = () => {
  return useQuery({
    queryKey: todaySummaryQueryKeys.current(),
    queryFn: getTodaySummary,
    retry: 2,
  });
};
