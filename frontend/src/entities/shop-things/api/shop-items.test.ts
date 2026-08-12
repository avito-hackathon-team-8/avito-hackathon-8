import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getShopItems, purchaseShopItem } from './shop-items';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('shop items API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue({ items: [] });
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('получает товары магазина с авторизацией', async () => {
    await expect(getShopItems()).resolves.toEqual([]);

    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/shop$/), {
      headers: { Authorization: 'Bearer token-1' },
    });
  });

  it('покупает товар POST-запросом с подтверждением замены', async () => {
    mocks.apiRequest.mockResolvedValue(undefined);

    await expect(
      purchaseShopItem('fashionable-bowl', { confirmReplacement: false }),
    ).resolves.toBeUndefined();

    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/shop\/fashionable-bowl\/purchase$/),
      {
        method: 'POST',
        headers: {
          Authorization: 'Bearer token-1',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ confirmReplacement: false }),
      },
    );
  });
});
