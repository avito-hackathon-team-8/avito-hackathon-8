import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getLevelsRewards, receiveLevelReward } from './levels-rewards';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('levels-rewards API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue({ levels: [] });
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('получает список уровней с AbortSignal', async () => {
    const controller = new AbortController();

    await getLevelsRewards(controller.signal);
    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/pet\/levels$/), {
      headers: { Authorization: 'Bearer token-1' },
      signal: controller.signal,
    });
  });

  it('получает награду уровня POST-запросом с AbortSignal', async () => {
    const controller = new AbortController();

    await receiveLevelReward('reward-3', controller.signal);
    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/pet\/level-rewards\/reward-3\/claim$/),
      {
        method: 'POST',
        headers: { Authorization: 'Bearer token-1' },
        signal: controller.signal,
      },
    );
  });
});
