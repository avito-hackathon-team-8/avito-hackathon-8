import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import type { TaskStatus, TTask } from '../../api/tasks';

import { DailyTaskPanelContent } from './daily-task-panel-content';

const createTask = (taskId: string, status: TaskStatus): TTask => ({
  taskId,
  slot: Number(taskId.at(-1)),
  type: 'VIEW_LISTINGS',
  description: `Задание ${taskId}`,
  currentCount: status === 'COMPLETED' || status === 'CLAIMED' ? 5 : 2,
  targetCount: 5,
  rewardLeaves: 20,
  requiredLevel: 3,
  status,
});

describe('DailyTaskPanelContent', () => {
  it('показывает состояния заданий и разрешает забрать выполненное', async () => {
    const user = userEvent.setup();
    const handleReceiveReward = vi.fn();
    const tasks = [
      createTask('task-1', 'IN_PROGRESS'),
      createTask('task-2', 'COMPLETED'),
      createTask('task-3', 'CLAIMED'),
      createTask('task-4', 'LOCKED'),
    ];

    render(<DailyTaskPanelContent listTasks={tasks} handleReceiveReward={handleReceiveReward} />);

    expect(screen.getByText('2 из 5')).toBeInTheDocument();
    expect(screen.getByText('Выполнено')).toBeInTheDocument();
    expect(screen.getByText('Закрыто')).toBeInTheDocument();
    expect(screen.getByText('Разблокируется на 3 уровне')).toBeInTheDocument();

    const rewardButtons = screen.getAllByRole('button', { name: 'Забрать награду' });
    expect(rewardButtons.filter((button) => !button.hasAttribute('disabled'))).toHaveLength(1);

    await user.click(rewardButtons[1]);
    expect(handleReceiveReward).toHaveBeenCalledWith('task-2');
  });

  it('не вызывает обработчик для незавершённого задания', async () => {
    const user = userEvent.setup();
    const handleReceiveReward = vi.fn();

    render(
      <DailyTaskPanelContent
        listTasks={[createTask('task-1', 'IN_PROGRESS')]}
        handleReceiveReward={handleReceiveReward}
      />,
    );

    const button = screen.getByRole('button', { name: 'Забрать награду' });
    expect(button).toBeDisabled();
    await user.click(button);
    expect(handleReceiveReward).not.toHaveBeenCalled();
  });
});
