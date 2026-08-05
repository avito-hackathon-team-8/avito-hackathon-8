import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import { useLeaderboard } from "../../model/use-leaderboard";
import pedestalIcon from "../assets/pedestal-icon.svg";
import { LeaderboardPanelContent } from "../leaderboard-panel-content/leaderboard-panel-content";

const TITLE_CARD = "Лидерборд";

export const LeaderboardCard = () => {
  const { data, refetch } = useLeaderboard();

  const handleClick = () => {
    if (!data) {
      refetch();
    }
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      description="Топ-10 платформы. Обновляется раз в день"
      disabled={!data}
      onClick={handleClick}
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
      {data && <LeaderboardPanelContent listUsers={data} />}
    </BottomPanel>
  );
};
