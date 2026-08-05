import { activityDaysMock } from "../model/mock/activity-day-mock";

export type TActivityDayStatus = "CLAIMED" | "MISSED" | "AVAILABLE" | "FUTURE";

type TWeekDay = 1 | 2 | 3 | 4 | 5 | 6 | 7;

export type TActivityDay = {
  weekday: TWeekDay;
  date: string;
  status: TActivityDayStatus;
  rewardLeaves: number;
  claimId?: string;
};

export type TResponseActivityDay = {
  claimedDaysCount: TWeekDay;
  claims: TActivityDay[];
};

export const getActivityDay = async (): Promise<TResponseActivityDay> => {
  return await Promise.resolve(activityDaysMock);
};
