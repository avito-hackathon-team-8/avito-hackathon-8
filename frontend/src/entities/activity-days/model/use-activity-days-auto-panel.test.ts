import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useActivityDaysAutoPanel } from './use-activity-days-auto-panel';

const mocks = vi.hoisted(() => ({
  useActivityDays: vi.fn(),
  useRecordTodayActivity: vi.fn(),
  receiveReward: vi.fn(),
}));

vi.mock('./use-activity-days', () => ({
  useActivityDays: mocks.useActivityDays,
}));

vi.mock('./use-record-today-activity', () => ({
  useRecordTodayActivity: mocks.useRecordTodayActivity,
}));

const week = {
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

describe('useActivityDaysAutoPanel', () => {
  beforeEach(() => {
    mocks.useActivityDays.mockReset().mockReturnValue({
      data: week,
      receiveReward: mocks.receiveReward,
    });
    mocks.useRecordTodayActivity.mockReset();
    mocks.receiveReward.mockReset();
  });

  it('записывает активность и читает только кэш недели', () => {
    renderHook(() => useActivityDaysAutoPanel());

    expect(mocks.useRecordTodayActivity).toHaveBeenCalledOnce();
    expect(mocks.useActivityDays).toHaveBeenCalledWith({ enabled: false });
  });

  it('открывает панель только при наличии доступного дня', () => {
    const { result, rerender } = renderHook(() => useActivityDaysAutoPanel());
    expect(result.current.isOpen).toBe(true);

    mocks.useActivityDays.mockReturnValue({
      data: {
        ...week,
        claims: week.claims.map((claim) => ({ ...claim, status: 'CLAIMED' })),
      },
      receiveReward: mocks.receiveReward,
    });
    rerender();
    expect(result.current.isOpen).toBe(false);
  });

  it('передаёт получение награды в useActivityDays', () => {
    const { result } = renderHook(() => useActivityDaysAutoPanel());

    act(() => result.current.handleReceiveReward());
    expect(mocks.receiveReward).toHaveBeenCalledOnce();
  });
});
