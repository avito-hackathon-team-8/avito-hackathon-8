export const activityDayKeys = {
  all: ["activity-days"] as const,

  current: () => [...activityDayKeys.all, "current"] as const,
};
