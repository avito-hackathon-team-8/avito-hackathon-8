import { BottomPanel } from "@/shared/ui/bottom-panel";
import { GamificationCard } from "@/shared/ui/gamification-card";

import rulesIcon from "../assets/rules-icon.svg";

const TITLE_CARD = "Правила";

type TRulesCardProps = {
  className?: string;
};

export const RulesCard = ({ className }: TRulesCardProps) => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      renderTrigger={(open) => (
        <GamificationCard
          variant="horizontal"
          title={TITLE_CARD}
          description="Как работают листья, уровни и награды"
          imageProps={{
            src: rulesIcon,
            alt: "Книжка правил",
            width: 94,
            height: 74,
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
