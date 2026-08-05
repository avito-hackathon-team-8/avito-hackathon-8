import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import chartIcon from "../assets/chart-icon.svg";

const TITLE_CARD = "Сводка дня";

type TTodaySummaryCard = {
  className?: string;
};

export const TodaySummaryCard = ({ className }: TTodaySummaryCard) => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      renderTrigger={(open) => (
        <GamificationCard
          variant="horizontal"
          title={TITLE_CARD}
          description="3 задания · +140 листьев · место 18"

          imageProps={{
            src: chartIcon,
            alt: "Календарь задач",
            width: 100,
            height: 94,
          }}
          wrapperProps={{ onClick: open }}
          className={className}
        />
      )}
    >
      <div></div>
    </BottomPanel>
  );
};
