import type { TResponseActivityDay } from "../../api/get-activity-day";

export const activityDaysMock: TResponseActivityDay = {
  claimedDaysCount: 1,
  claims: [
    {
      weekday: 1,
      date: "2026-08-03",
      status: "CLAIMED",
      rewardLeaves: 10,
      claimId: "40f36ae1-70a3-447e-b418-d5eac91e0927",
    },
    {
      weekday: 2,
      date: "2026-08-04",
      status: "MISSED",
      rewardLeaves: 20,
    },
    {
      weekday: 3,
      date: "2026-08-05",
      status: "AVAILABLE",
      rewardLeaves: 30,
    },
    {
      weekday: 4,
      date: "2026-08-06",
      status: "FUTURE",
      rewardLeaves: 40,
    },
    {
      weekday: 5,
      date: "2026-08-07",
      status: "FUTURE",
      rewardLeaves: 50,
    },
    {
      weekday: 6,
      date: "2026-08-08",
      status: "FUTURE",
      rewardLeaves: 60,
    },
    {
      weekday: 7,
      date: "2026-08-09",
      status: "FUTURE",
      rewardLeaves: 70,
    },
  ],
};
