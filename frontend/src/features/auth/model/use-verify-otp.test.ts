import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { mainQueryKey } from '@/shared/config';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { useVerifyOtp } from './use-verify-otp';

const mocks = vi.hoisted(() => ({
  verifyOtp: vi.fn(),
  getCurrentUser: vi.fn(),
  setSessionStorageValue: vi.fn(),
}));

vi.mock('../api/auth', () => ({
  verifyOtp: mocks.verifyOtp,
}));

vi.mock('@/entities/user', () => ({
  getCurrentUser: mocks.getCurrentUser,
  userQueryKeys: {
    current: () => ['user', 'current'],
  },
}));

vi.mock('@/shared/lib', () => ({
  sessionStorageKeysMap: { authToken: 'auth_token' },
  setSessionStorageValue: mocks.setSessionStorageValue,
}));

describe('useVerifyOtp', () => {
  beforeEach(() => {
    mocks.verifyOtp.mockReset();
    mocks.getCurrentUser.mockReset();
    mocks.setSessionStorageValue.mockReset();
  });

  it('сохраняет токен, загружает пользователя и обновляет кэш приложения', async () => {
    const authResponse = {
      token: 'token-1',
      record: { id: 'user-1', email: 'user@example.com', verified: true },
    };
    const currentUser = { id: 'user-1', email: 'user@example.com', verified: true };
    mocks.verifyOtp.mockResolvedValue(authResponse);
    mocks.getCurrentUser.mockResolvedValue(currentUser);

    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useVerifyOtp(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ email: 'user@example.com', code: '12345678' });
    });

    expect(mocks.verifyOtp).toHaveBeenCalledWith('user@example.com', '12345678');
    expect(mocks.setSessionStorageValue).toHaveBeenCalledWith('auth_token', 'token-1');
    expect(mocks.getCurrentUser).toHaveBeenCalledWith('token-1');
    expect(queryClient.getQueryData(['user', 'current'])).toEqual(currentUser);
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: mainQueryKey.all });
  });

  it('не выполняет побочные эффекты при ошибке подтверждения', async () => {
    mocks.verifyOtp.mockRejectedValue(new Error('Неверный код'));
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useVerifyOtp(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => {
      await expect(
        result.current.mutateAsync({ email: 'user@example.com', code: '00000000' }),
      ).rejects.toThrow('Неверный код');
    });

    expect(mocks.setSessionStorageValue).not.toHaveBeenCalled();
    expect(mocks.getCurrentUser).not.toHaveBeenCalled();
  });
});
