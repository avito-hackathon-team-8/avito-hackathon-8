import { useCallback } from 'react';

import { useActivityDays } from './use-activity-days';
import { useRecordTodayActivity } from './use-record-today-activity';

export const useActivityDaysAutoPanel = () => {
  useRecordTodayActivity();
  const { data: cachedActivityDays, receiveReward } = useActivityDays({
    enabled: false,
  });

  const data = cachedActivityDays;

  const hasAvailableDay = data?.claims.some(({ status }) => status === 'AVAILABLE');
  const isOpen = Boolean(hasAvailableDay);

  const handleReceiveReward = useCallback(() => {
    receiveReward();
  }, [receiveReward]);

  return {
    data,
    isOpen,
    handleReceiveReward,
  };
};
