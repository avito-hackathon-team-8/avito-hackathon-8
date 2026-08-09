import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_DAILY_TASKS } from './api-routes';

export type TActivityDayStatus = 'CLAIMED' | 'MISSED' | 'AVAILABLE' | 'FUTURE';

type TWeekDay = 1 | 2 | 3 | 4 | 5 | 6 | 7;

export type TActivityDay = {
  weekday: TWeekDay;
  date: string;
  status: TActivityDayStatus;
  rewardLeaves: number;
  claimId: string;
};

export type TResponseActivityDay = {
  claimedDaysCount: TWeekDay;
  claims: TActivityDay[];
};

export const getActivityDay = async (): Promise<TResponseActivityDay> => {
  return await apiRequest(
    fetch(`${API_URL}${API_ROUTE_DAILY_TASKS.week}`, {
      headers: getAuthHeaders(),
    }),
  );
};

export const receiveActivityDayReward = async (): Promise<void> => {
  return apiRequest<void>(
    fetch(`${API_URL}${API_ROUTE_DAILY_TASKS.receiveRewardDay()}`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
  );
};

export const recordTodayActivity = async (): Promise<void> => {
  return apiRequest<void>(
    fetch(`${API_URL}${API_ROUTE_DAILY_TASKS.activity}`, {
      method: 'POST',
      headers: getAuthHeaders(),
    }),
  );
};
