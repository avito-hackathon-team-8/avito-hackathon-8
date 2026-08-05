import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import tasksBoardIcon from "../assets/tasks-board-icon.svg";

const TITLE_CARD = "Ежедневные задания";

export const DailyTaskCard = () => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description="2 из 4 выполнено"
          imageProps={{
            src: tasksBoardIcon,
            alt: "Доска задач",
            width: 78,
            height: 86,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      <div></div>
    </BottomPanel>
  );
};
