import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getTasks, receiveTaskReward } from './tasks';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

describe('tasks API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset().mockResolvedValue({ tasks: [] });
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('получает ежедневные задания с авторизацией', async () => {
    await getTasks();

    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/tasks$/), {
      headers: { Authorization: 'Bearer token-1' },
    });
  });

  it('получает награду задания POST-запросом', async () => {
    await receiveTaskReward('task-1');

    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/v1\/tasks\/task-1\/claim$/),
      {
        headers: { Authorization: 'Bearer token-1' },
        method: 'POST',
      },
    );
  });
});
