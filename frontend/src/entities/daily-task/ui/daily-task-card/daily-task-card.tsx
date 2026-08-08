import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import type { TTask } from '../../api/tasks';
import { useDailyTasks } from '../../model/use-daily-tasks';
import tasksBoardIcon from '../assets/tasks-board-icon.svg';
import { DailyTaskPanelContent } from '../daily-task-panel-content/daily-task-panel-content';

const TITLE_CARD = 'Ежедневные задания';

export const DailyTaskCard = () => {
  const { data, isLoading, receiveReward } = useDailyTasks();

  if (!data) return;

  const { tasks } = data;

  const countCompeteTask = tasks.reduce((acc, item) => {
    if (item.status === 'CLAIMED' || item.status === 'COMPLETED') {
      const count = acc + 1;
      return count;
    }

    return acc;
  }, 0);

  const handleReceiveRewardClick = (taskId: TTask['taskId']) => {
    receiveReward({ taskId });
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      disabled={isLoading}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={`${countCompeteTask} из ${tasks?.length} выполнено`}
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
      <DailyTaskPanelContent listTasks={tasks} handleReceiveReward={handleReceiveRewardClick} />
    </BottomPanel>
  );
};
