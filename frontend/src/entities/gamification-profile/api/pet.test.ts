import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getPet, updatePetName } from './pet';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  getAuthHeaders: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
  getAuthHeaders: mocks.getAuthHeaders,
}));

const pet = {
  name: 'Листик',
  level: 3,
  leaves: 300,
  nextLevelTargetLeaves: 500,
  chestPrice: 100,
  levelUp: false,
  bowlImageUrl: null,
  bedImageUrl: null,
  happiness: 100,
  happinessMultiplier: 1.5,
  calculatedAt: '2026-08-12T12:52:25.179950567Z',
  decaysToZeroAt: '2026-08-15T12:52:15.223227999Z',
  feedNextAvailableAt: null,
  strokeNextAvailableAt: null,
};

describe('pet API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset();
    mocks.apiRequest.mockReset().mockResolvedValue(pet);
    mocks.getAuthHeaders.mockReset().mockReturnValue({ Authorization: 'Bearer token-1' });
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('получает профиль питомца', async () => {
    mocks.fetch.mockResolvedValue(
      new Response(JSON.stringify(pet), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    await expect(getPet()).resolves.toEqual(pet);
    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/pet$/), {
      headers: { Authorization: 'Bearer token-1' },
    });
  });

  it('выбрасывает ошибку со статусом неуспешного запроса профиля', async () => {
    mocks.fetch.mockResolvedValue(new Response(null, { status: 503 }));

    await expect(getPet()).rejects.toThrow('Ошибка запроса getPet: 503');
  });

  it('обновляет имя питомца PATCH-запросом', async () => {
    mocks.fetch.mockResolvedValue(new Response());

    await expect(updatePetName('Коробыш')).resolves.toEqual(pet);
    expect(mocks.fetch).toHaveBeenCalledWith(expect.stringMatching(/\/api\/v1\/pet$/), {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer token-1',
      },
      body: JSON.stringify({ name: 'Коробыш' }),
    });
    expect(mocks.apiRequest).toHaveBeenCalledWith(
      expect.any(Promise),
      'Не удалось сохранить имя питомца',
    );
  });
});
