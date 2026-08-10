import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { openChest } from './open-chest';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('openChest', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue({ id: 'reward-1' });
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('открывает сундук авторизованным POST-запросом', async () => {
    await expect(openChest()).resolves.toEqual({ id: 'reward-1' });
    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/pet\/chests\/open$/),
      {
        method: 'POST',
        headers: { Authorization: 'Bearer token-1' },
      },
    );
    expect(mocks.apiRequest).toHaveBeenCalledWith(expect.any(Promise), 'Не удалось открыть сундук');
  });
});
