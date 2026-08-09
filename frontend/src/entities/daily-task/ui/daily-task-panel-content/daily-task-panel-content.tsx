import clsx from 'clsx';

import { leafIcon } from '@/shared/assets/icon';
import { Typography } from '@/shared/ui/typography';

import type { TTask } from '../../api/tasks';

import styles from './daily-task-panel-content.module.scss';

interface IDailyTaskPanelContentProps {
  listTasks: TTask[];
  handleReceiveReward: (id: TTask['taskId']) => void;
}

export const DailyTaskPanelContent = ({
  listTasks,
  handleReceiveReward,
}: IDailyTaskPanelContentProps) => {
  return (
    <ul className={styles.listTask}>
      {listTasks.map((item) => {
        const isLocked = item.status === 'LOCKED';
        const isClaimed = item.status === 'CLAIMED';
        const isCompleted = item.status === 'COMPLETED';

        return (
          <li
            className={clsx(styles.listTask__item, {
              [styles.listTask__item_locked]: isLocked,
              [styles.listTask__item_completed]: isCompleted,
              [styles.listTask__item_claimed]: isClaimed,
            })}
            key={item.taskId}
          >
            <Typography className={styles.listTask__itemNumber} variant="p3" color="inherit">
              {!isLocked && item.slot}
              {isLocked && '×'}
              {}
            </Typography>

            <div className={styles.listTask__content}>
              <Typography
                className={styles.listTask__contentTitle}
                color="inherit"
                variant="caption-semiBold"
                as="h3"
              >
                {item.description}
              </Typography>

              <Typography variant="p4-regular" color="gray500">
                {!isLocked && `${item.currentCount} из ${item.targetCount}`}
                {isLocked && `Разблокируется на ${item.requiredLevel} уровне`}
              </Typography>
            </div>

            <div className={styles.listTask__info}>
              <Typography className={styles.listTask__reward} variant="caption-semiBold">
                <img width={24} height={24} src={leafIcon} aria-hidden />
                <span
                  aria-label="Кол-во листьев за выполнение задания"
                  className={styles.listTask__rewardText}
                >
                  {item.rewardLeaves}
                </span>
              </Typography>

              <Typography variant="p4-regular" color="gray500">
                {isLocked && 'Закрыто'}
                {isCompleted && 'Выполнено'}
              </Typography>
            </div>

            <button
              className={styles.listTask__button}
              aria-label="Забрать награду"
              disabled={!isCompleted}
              onClick={() => handleReceiveReward(item.taskId)}
            />
          </li>
        );
      })}
    </ul>
  );
};
