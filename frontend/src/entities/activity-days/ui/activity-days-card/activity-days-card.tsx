import { formatDays } from '@/shared/lib';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useActivityDays } from '../../model/use-activity-days';
import { ActivityDaysPanelContent } from '../activity-days-panel-content/activity-days-panel-content';
import calendarIcon from '../assets/calendar-icon.webp';

const TITLE_CARD = 'Дни активности';
const DESCRIPTION_ERROR = 'Не удалось получить данные';

export const ActivityDaysCard = () => {
  const { data, refetch, isPending, receiveReward } = useActivityDays({
    enabled: false,
  });

  const handleClick = () => {
    if (!data && !isPending) {
      refetch();
    }
  };

  const handleReceiveRewardClick = () => {
    receiveReward();
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      description="Награда зависит от календарного дня"
      disabled={!data}
      onClick={handleClick}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={
            data ? `${formatDays(data.claimedDaysCount)} на этой неделе` : DESCRIPTION_ERROR
          }
          imageProps={{
            src: calendarIcon,
            alt: 'Календарь задач',
            width: 70,
            height: 72,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {data && (
        <ActivityDaysPanelContent data={data} handleReceiveReward={handleReceiveRewardClick} />
      )}
    </BottomPanel>
  );
};
