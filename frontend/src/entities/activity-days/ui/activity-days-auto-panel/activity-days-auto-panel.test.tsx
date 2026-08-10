import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ActivityDaysAutoPanel } from './activity-days-auto-panel';

const mocks = vi.hoisted(() => ({
  useActivityDaysAutoPanel: vi.fn(),
  handleReceiveReward: vi.fn(),
}));

vi.mock('../../model/use-activity-days-auto-panel', () => ({
  useActivityDaysAutoPanel: mocks.useActivityDaysAutoPanel,
}));

const data = {
  claimedDaysCount: 2,
  claims: [
    {
      weekday: 2,
      date: '2026-08-11',
      status: 'AVAILABLE',
      rewardLeaves: 20,
      claimId: 'claim-2',
    },
  ],
};

describe('ActivityDaysAutoPanel', () => {
  beforeEach(() => {
    const portalRoot = document.createElement('div');
    portalRoot.id = 'app-overlay-root';
    document.body.append(portalRoot);

    mocks.useActivityDaysAutoPanel.mockReset();
    mocks.handleReceiveReward.mockReset();
  });

  afterEach(() => {
    document.getElementById('app-overlay-root')?.remove();
    document.documentElement.style.overflow = '';
    document.body.style.overflow = '';
  });

  it('ничего не показывает без доступного дня', () => {
    mocks.useActivityDaysAutoPanel.mockReturnValue({
      data,
      isOpen: false,
      handleReceiveReward: mocks.handleReceiveReward,
    });

    const { container } = render(<ActivityDaysAutoPanel />);
    expect(container).toBeEmptyDOMElement();
  });

  it('автоматически открывает панель и позволяет забрать награду', async () => {
    const user = userEvent.setup();
    mocks.useActivityDaysAutoPanel.mockReturnValue({
      data,
      isOpen: true,
      handleReceiveReward: mocks.handleReceiveReward,
    });

    render(<ActivityDaysAutoPanel />);

    expect(screen.getByRole('dialog', { name: 'Дни активности' }).parentElement).toHaveAttribute(
      'data-open',
      'true',
    );
    await user.click(screen.getByRole('button', { name: 'вт 20' }));
    expect(mocks.handleReceiveReward).toHaveBeenCalledOnce();
  });
});
