import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import calendarIcon from "../assets/calendar-icon.svg";

const TITLE_CARD = "Дни активности";

export const ActivityDaysCard = () => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description="5 дней на этой неделе"
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
      <div></div>
    </BottomPanel>
  );
};
