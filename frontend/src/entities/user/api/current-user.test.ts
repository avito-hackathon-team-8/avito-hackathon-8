import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { sessionStorageKeysMap } from '@/shared/lib';

import { getCurrentUser } from './current-user';

const fetchMock = vi.fn();

describe('getCurrentUser', () => {
  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('получает пользователя с Authorization header', async () => {
    const user = { id: 'user-1', email: 'user@example.com', verified: true };
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify(user), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    await expect(getCurrentUser('token-1')).resolves.toEqual(user);
    expect(fetchMock).toHaveBeenCalledWith(expect.stringMatching(/\/api\/app\/auth\/me$/), {
      headers: { Authorization: 'Bearer token-1' },
    });
  });

  it('удаляет токен, если сервер отклонил авторизацию', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'bad-token');
    fetchMock.mockResolvedValue(new Response('Unauthorized', { status: 401 }));

    await expect(getCurrentUser('bad-token')).rejects.toThrow('Unauthorized');
    expect(sessionStorage.getItem(sessionStorageKeysMap.authToken)).toBeNull();
  });
});
