import { formatWord } from '@/shared/lib';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useTodaySummary } from '../../model/use-today-summary';
import chartIcon from '../assets/chart-icon.webp';
import { TodaySummaryEmpty } from '../today-summary-empty/today-summary-empty';
import { TodaySummaryPanelContent } from '../today-summary-panel-content/today-summary-panel-content';

const TITLE_CARD = 'Сводка дня';

type TTodaySummaryCard = {
  className?: string;
};

const WORD_ACTIVITY = ['Активность', 'Активности', 'Активностей'] as const;

export const TodaySummaryCard = ({ className }: TTodaySummaryCard) => {
  const { data, refetch } = useTodaySummary();

  const activitiesCount = data?.activitiesCount || 0;

  return (
    <BottomPanel
      title={TITLE_CARD}
      onClick={() => {
        if (!data) {
          refetch();
        }
      }}
      disabled={!data}
      renderTrigger={(open) => (
        <GamificationCard
          variant="horizontal"
          title={TITLE_CARD}
          description={`${formatWord(activitiesCount, WORD_ACTIVITY)}: ${activitiesCount} · +${data?.leavesCount} листьев`}

          imageProps={{
            src: chartIcon,
            alt: 'Календарь задач',
            width: 84,
            height: 73,
          }}
          wrapperProps={{ onClick: open }}
          className={className}
        />
      )}
    >
      {data && data.events.length !== 0 && <TodaySummaryPanelContent events={data} />}
      {data && data.events.length === 0 && <TodaySummaryEmpty />}
    </BottomPanel>
  );
};
