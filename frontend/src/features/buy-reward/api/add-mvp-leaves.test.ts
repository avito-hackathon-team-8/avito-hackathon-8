import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { addMVPLeaves } from './add-mvp-leaves';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('addMVPLeaves', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue({ leaves: 700 });
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('начисляет листья авторизованным POST-запросом', async () => {
    await expect(addMVPLeaves()).resolves.toEqual({ leaves: 700 });
    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/pet\/mvp\/leaves$/),
      {
        method: 'POST',
        headers: { Authorization: 'Bearer token-1' },
      },
    );
    expect(mocks.apiRequest).toHaveBeenCalledWith(
      expect.any(Promise),
      'Не удалось начислить листья',
    );
  });
});
