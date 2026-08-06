import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useDailyTasks } from '../../model/use-daily-tasks';
import tasksBoardIcon from '../assets/tasks-board-icon.svg';
import { DailyTaskPanelContent } from '../daily-task-panel-content/daily-task-panel-content';

const TITLE_CARD = 'Ежедневные задания';

export const DailyTaskCard = () => {
  const { data } = useDailyTasks();

  if (!data) return;

  const countCompeteTask = data.reduce((acc, item) => {
    if (item.status === 'COMPLETED') {
      const count = acc + 1;
      return count;
    }

    return acc;
  }, 0);

  return (
    <BottomPanel
      title={TITLE_CARD}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={`${countCompeteTask} из ${data?.length} выполнено`}
          imageProps={{
            src: tasksBoardIcon,
            alt: 'Доска задач',
            width: 78,
            height: 86,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      <DailyTaskPanelContent listTasks={data} />
    </BottomPanel>
  );
};
