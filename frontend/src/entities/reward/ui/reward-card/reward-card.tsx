import { useMemo } from 'react';

import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useRewards } from '../../model/use-reward';
import cupIcon from '../assets/сup-icon.svg';
import { RewardsPanelContent } from '../rewards-panel-content/rewards-panel-content';

const TITLE_CARD = 'Награды';

export const RewardCard = () => {
  const { data: rewards, isPending, refetch } = useRewards();

  const listReward = useMemo(() => {
    return rewards?.groups.flatMap((item) => item.items);
  }, [rewards]);

  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      disabled={!listReward || listReward.length === 0}
      onClick={() => {
        if (!rewards && !isPending) {
          refetch();
        }
      }}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={`${listReward?.length || 0} бонуса доступны`}
          imageProps={{ src: cupIcon, alt: 'кубок', width: 92, height: 80 }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {rewards && rewards.groups.length > 0 && listReward && (
        <RewardsPanelContent listReward={listReward} />
      )}
    </BottomPanel>
  );
};
