import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getActivityDay, receiveActivityDayReward, recordTodayActivity } from './activity-day';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('activity-day API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue(undefined);
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('запрашивает данные недели с авторизацией', async () => {
    await getActivityDay();

    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/weekly-login$/), {
      headers: { Authorization: 'Bearer token-1' },
    });
  });

  it('отправляет запрос на получение награды', async () => {
    await receiveActivityDayReward();

    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/weekly-login\/claim$/),
      {
        method: 'POST',
        headers: { Authorization: 'Bearer token-1' },
      },
    );
  });

  it('отправляет запрос на запись сегодняшней активности', async () => {
    await recordTodayActivity();

    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/weekly-login\/activity$/),
      {
        method: 'POST',
        headers: { Authorization: 'Bearer token-1' },
      },
    );
  });
});
