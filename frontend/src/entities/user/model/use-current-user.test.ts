import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { sessionStorageKeysMap } from '@/shared/lib';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { useCurrentUser } from './use-current-user';

const mocks = vi.hoisted(() => ({
  getCurrentUser: vi.fn(),
}));

vi.mock('../api/current-user', () => ({
  getCurrentUser: mocks.getCurrentUser,
}));

describe('useCurrentUser', () => {
  beforeEach(() => {
    mocks.getCurrentUser.mockReset();
  });

  it('не запрашивает пользователя без токена', () => {
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useCurrentUser(), {
      wrapper: createQueryWrapper(queryClient),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mocks.getCurrentUser).not.toHaveBeenCalled();
  });

  it('запрашивает пользователя с токеном из sessionStorage', async () => {
    const currentUser = { id: 'user-1', email: 'user@example.com', verified: true };
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    mocks.getCurrentUser.mockResolvedValue(currentUser);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useCurrentUser(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mocks.getCurrentUser).toHaveBeenCalledWith('token-1');
    expect(result.current.data).toEqual(currentUser);
  });
});
