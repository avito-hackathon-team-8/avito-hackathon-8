import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import pedestalIcon from "../assets/pedestal-icon.svg";

const TITLE_CARD = "Лидерборд";

export const LeaderboardCard = () => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description="Ваше место: 18"
          imageProps={{
            src: pedestalIcon,
            alt: "Пьедестал",
            width: 96,
            height: 65,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      <div></div>
    </BottomPanel>
  );
};
