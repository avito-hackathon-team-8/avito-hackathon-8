import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import type { TTask } from '../../api/tasks';
import { useDailyTasks } from '../../model/use-daily-tasks';
import tasksBoardIcon from '../assets/tasks-board-icon.webp';
import { DailyTaskPanelContent } from '../daily-task-panel-content/daily-task-panel-content';

const TITLE_CARD = 'Ежедневные задания';
const TASKS_EMPTY_TEXT = 'На данный момент задач нету';

export const DailyTaskCard = () => {
  const { data: tasks, isPending, refetch, receiveReward } = useDailyTasks();

  const countCompeteTask =
    tasks &&
    tasks.reduce((acc, item) => {
      if (item.status === 'CLAIMED' || item.status === 'COMPLETED') {
        const count = acc + 1;
        return count;
      }

      return acc;
    }, 0);

  const description =
    countCompeteTask && tasks
      ? `${countCompeteTask} из ${tasks.length} выполнено`
      : TASKS_EMPTY_TEXT;

  const handleReceiveRewardClick = (taskId: TTask['taskId']) => {
    receiveReward({ taskId });
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      disabled={!tasks}
      onClick={() => {
        if (!tasks && !isPending) {
          refetch();
        }
      }}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={description}
          imageProps={{
            src: tasksBoardIcon,
            alt: 'Доска задач',
            width: 69,
            height: 74,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {tasks && (
        <DailyTaskPanelContent listTasks={tasks} handleReceiveReward={handleReceiveRewardClick} />
      )}
    </BottomPanel>
  );
};
