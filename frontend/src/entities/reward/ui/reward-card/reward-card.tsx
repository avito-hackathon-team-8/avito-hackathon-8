import { useMemo } from 'react';

import { formatWord } from '@/shared/lib';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useRewards } from '../../model/use-reward';
import cupIcon from '../assets/сup-icon.webp';
import { RewardsPanelContent } from '../rewards-panel-content/rewards-panel-content';

const TITLE_CARD = 'Награды';

const BONUS_TEXT = ['бонус', 'бонуса', 'бонусов'] as const;
const REWARDS_COUNT = ['доступен', 'доступны', 'доступно'] as const;

export const RewardCard = () => {
  const { data: rewards, isPending, refetch } = useRewards();

  const listReward = useMemo(() => {
    return rewards?.groups.flatMap((item) => item.items);
  }, [rewards]);
  const rewardsCount = listReward?.length || 0;

  const getDescription = `${rewardsCount} ${formatWord(rewardsCount, BONUS_TEXT)} ${formatWord(
    rewardsCount,
    REWARDS_COUNT,
  )}`;

  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      disabled={!listReward}
      onClick={() => {
        if (!rewards && !isPending) {
          refetch();
        }
      }}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={getDescription}
          imageProps={{ src: cupIcon, alt: 'кубок', width: 63, height: 75 }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {rewards && listReward && <RewardsPanelContent listReward={listReward} />}
    </BottomPanel>
  );
};
