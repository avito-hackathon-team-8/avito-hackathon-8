import { formatDays } from "@/shared/lib/format-days";
import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import { useActivityDays } from "../../model/use-activity-days";
import { ActivityDaysPanelContent } from "../activity-days-panel-content/activity-days-panel-content";
import calendarIcon from "../assets/calendar-icon.svg";

const TITLE_CARD = "Дни активности";
const DESCRIPTION_ERROR = "Не удалось получить данные";

export const ActivityDaysCard = () => {
  const { data, refetch } = useActivityDays();

  const handleClick = () => {
    if (!data) {
      refetch();
    }
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      description="Награда зависит от календарного дня"
      disabled={!data}
      onClick={handleClick}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={
            data
              ? `${formatDays(data.claimedDaysCount)} на этой неделе`
              : DESCRIPTION_ERROR
          }
          imageProps={{
            src: calendarIcon,
            alt: "Календарь задач",
            width: 84,
            height: 76,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {data && <ActivityDaysPanelContent data={data} />}
    </BottomPanel>
  );
};
