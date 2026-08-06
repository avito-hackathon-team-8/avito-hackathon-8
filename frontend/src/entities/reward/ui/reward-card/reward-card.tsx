import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { mockRewards } from '../../model/mock/mockRewards';
import cupIcon from '../assets/сup-icon.svg';
import { RewardsPanelContent } from '../rewards-panel-content/rewards-panel-content';

const TITLE_CARD = 'Награды';

export const RewardCard = () => {
  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description="2 из 4 выполнено"
          imageProps={{ src: cupIcon, alt: 'кубок', width: 92, height: 80 }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      <RewardsPanelContent listReward={mockRewards[0].items} />
    </BottomPanel>
  );
};
