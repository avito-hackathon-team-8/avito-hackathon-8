import { useQuery } from '@tanstack/react-query';

import { getTodaySummary } from '../api/get-today-summary';
import { todaySummaryQueryKeys } from '../api/today-summary-query-keys';

import { getTodaySummaryStats } from './get-today-summary-stats ';

export const useTodaySummary = () => {
  return useQuery({
    queryKey: todaySummaryQueryKeys.current(),
    queryFn: getTodaySummary,
    select: getTodaySummaryStats,
    retry: 2,
  });
};
