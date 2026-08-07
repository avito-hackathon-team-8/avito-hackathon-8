import { formatTasks } from '@/shared/lib/format-tasks';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useTodaySummary } from '../../model/use-today-summary';
import chartIcon from '../assets/chart-icon.svg';
import { TodaySummaryPanelContent } from '../today-summary-panel-content/today-summary-panel-content';

const TITLE_CARD = 'Сводка дня';

type TTodaySummaryCard = {
  className?: string;
};

export const TodaySummaryCard = ({ className }: TTodaySummaryCard) => {
  const { data, refetch } = useTodaySummary();
  return (
    <BottomPanel
      title={TITLE_CARD}
      description="Ежедневная сводка"
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
          description={`${data?.activitiesCount || 0} ${formatTasks(data?.activitiesCount || 0)} · +${data?.leavesCount} листьев`}

          imageProps={{
            src: chartIcon,
            alt: 'Календарь задач',
            width: 100,
            height: 94,
          }}
          wrapperProps={{ onClick: open }}
          className={className}
        />
      )}
    >
      {data && <TodaySummaryPanelContent events={data} />}
    </BottomPanel>
  );
};
