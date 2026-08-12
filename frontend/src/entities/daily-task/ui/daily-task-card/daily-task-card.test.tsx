import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';

import type { TaskStatus, TTask } from '../../api/tasks';

import { DailyTaskCard } from './daily-task-card';

const mocks = vi.hoisted(() => ({
  useDailyTasks: vi.fn(),
  receiveReward: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('../../model/use-daily-tasks', () => ({
  useDailyTasks: mocks.useDailyTasks,
}));

type BottomPanelStubProps = Pick<BottomPanelProps, 'children' | 'disabled' | 'renderTrigger'>;

vi.mock('@/shared/ui/bottom-panel', () => ({
  BottomPanel: ({ children, disabled, renderTrigger }: BottomPanelStubProps) => (
    <div data-testid="bottom-panel" data-disabled={disabled}>
      {renderTrigger(() => undefined)}
      {children}
    </div>
  ),
}));

const createTask = (taskId: string, status: TaskStatus): TTask => ({
  taskId,
  slot: Number(taskId.at(-1)),
  type: 'VIEW_LISTINGS',
  description: `Задание ${taskId}`,
  currentCount: status === 'COMPLETED' || status === 'CLAIMED' ? 5 : 2,
  targetCount: 5,
  rewardLeaves: 20,
  requiredLevel: 1,
  status,
});

describe('DailyTaskCard', () => {
  beforeEach(() => {
    mocks.useDailyTasks.mockReset();
    mocks.receiveReward.mockReset();
    mocks.refetch.mockReset();
  });

  it('показывает заглушку и не открывается без данных', () => {
    mocks.useDailyTasks.mockReturnValue({
      data: undefined,
      isPending: true,
      refetch: mocks.refetch,
      receiveReward: mocks.receiveReward,
    });

    render(<DailyTaskCard />);

    expect(screen.getByText('Ежедневные задания')).toBeInTheDocument();
    expect(screen.getByText('На данный момент задач нету')).toBeInTheDocument();
    expect(screen.getByTestId('bottom-panel')).toHaveAttribute('data-disabled', 'true');
  });

  it('считает выполненные задания и передаёт получение награды', async () => {
    const user = userEvent.setup();
    mocks.useDailyTasks.mockReturnValue({
      data: [
        createTask('task-1', 'COMPLETED'),
        createTask('task-2', 'CLAIMED'),
        createTask('task-3', 'IN_PROGRESS'),
      ],
      isPending: false,
      refetch: mocks.refetch,
      receiveReward: mocks.receiveReward,
    });
    render(<DailyTaskCard />);

    expect(screen.getByText('2 из 3 выполнено')).toBeInTheDocument();
    await user.click(screen.getAllByRole('button', { name: 'Забрать награду' })[0]);
    expect(mocks.receiveReward).toHaveBeenCalledWith({ taskId: 'task-1' });
  });
});
