import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import type { TResponseActivityDay } from '../../api/activity-day';

import { ActivityDaysPanelContent } from './activity-days-panel-content';

const activityDays: TResponseActivityDay = {
  claimedDaysCount: 2,
  claims: [
    {
      weekday: 1,
      date: '2026-08-10',
      status: 'CLAIMED',
      rewardLeaves: 10,
      claimId: 'claim-1',
    },
    {
      weekday: 2,
      date: '2026-08-11',
      status: 'AVAILABLE',
      rewardLeaves: 20,
      claimId: 'claim-2',
    },
    {
      weekday: 3,
      date: '2026-08-12',
      status: 'FUTURE',
      rewardLeaves: 30,
      claimId: 'claim-3',
    },
  ],
};

describe('ActivityDaysPanelContent', () => {
  it('разрешает получить награду только за доступный день', async () => {
    const user = userEvent.setup();
    const handleReceiveReward = vi.fn();

    render(
      <ActivityDaysPanelContent data={activityDays} handleReceiveReward={handleReceiveReward} />,
    );

    expect(screen.getByText('Текущая серия: 2 дня')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'пн 10' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'вт 20' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'ср 30' })).toBeDisabled();

    await user.click(screen.getByRole('button', { name: 'вт 20' }));
    await user.click(screen.getByRole('button', { name: 'Забрать награду' }));
    expect(handleReceiveReward).toHaveBeenCalledTimes(2);
  });

  it('блокирует общую кнопку, если последний активный день недоступен', () => {
    render(
      <ActivityDaysPanelContent
        data={{
          claimedDaysCount: 1,
          claims: [{ ...activityDays.claims[0], status: 'CLAIMED' }],
        }}
        handleReceiveReward={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: 'Забрать награду' })).toBeDisabled();
  });
});
