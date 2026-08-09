import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { dataRules } from '../../model/rules';
import rulesIcon from '../assets/rules-icon.webp';
import { RulesPanelContent } from '../rules-panel-content/rules-panel-content';

const TITLE_CARD = 'Правила';

type TRulesCardProps = {
  className?: string;
};

export const RulesCard = ({ className }: TRulesCardProps) => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      renderTrigger={(open) => (
        <GamificationCard
          variant="horizontal"
          title={TITLE_CARD}
          description="Как работают листья, уровни и награды"
          imageProps={{
            src: rulesIcon,
            alt: 'Книжка правил',
            width: 65,
            height: 67,
          }}
          wrapperProps={{ onClick: open }}
          className={className}
        />
      )}
    >
      <RulesPanelContent rules={dataRules} />
    </BottomPanel>
  );
};
