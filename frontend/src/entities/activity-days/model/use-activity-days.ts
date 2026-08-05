import { useQuery } from "@tanstack/react-query";

import { activityDayKeys } from "../api/activity-days-keys";
import { getActivityDay } from "../api/get-activity-day";

export const useActivityDays = () => {
  return useQuery({
    queryKey: activityDayKeys.current(),
    queryFn: getActivityDay,
    retry: 2,
  });
};
