import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useOtpAuth } from './use-otp-auth';

const mutationMocks = vi.hoisted(() => ({
  requestOtp: vi.fn(),
  verifyOtp: vi.fn(),
  isRequesting: false,
  isVerifying: false,
}));

vi.mock('./use-request-otp', () => ({
  useRequestOtp: () => ({
    mutateAsync: mutationMocks.requestOtp,
    isPending: mutationMocks.isRequesting,
  }),
}));

vi.mock('./use-verify-otp', () => ({
  useVerifyOtp: () => ({
    mutateAsync: mutationMocks.verifyOtp,
    isPending: mutationMocks.isVerifying,
  }),
}));

describe('useOtpAuth', () => {
  beforeEach(() => {
    mutationMocks.requestOtp.mockReset().mockResolvedValue(undefined);
    mutationMocks.verifyOtp.mockReset().mockResolvedValue(undefined);
    mutationMocks.isRequesting = false;
    mutationMocks.isVerifying = false;
  });

  it('переходит с приветствия к вводу email', () => {
    const { result } = renderHook(() => useOtpAuth());

    expect(result.current.step).toBe('welcome');

    act(() => result.current.openEmailStep());

    expect(result.current.step).toBe('email');
    expect(result.current.error).toBe('');
  });

  it('проверяет email до запроса к API', async () => {
    const { result } = renderHook(() => useOtpAuth());

    act(() => result.current.changeEmail('not-an-email'));
    await act(async () => result.current.sendCode());

    expect(mutationMocks.requestOtp).not.toHaveBeenCalled();
    expect(result.current.error).toBe('Введите корректный email');
  });

  it('нормализует email и переходит к вводу кода после успешного запроса', async () => {
    const { result } = renderHook(() => useOtpAuth());

    act(() => result.current.changeEmail('  user@example.com  '));
    await act(async () => result.current.sendCode());

    expect(mutationMocks.requestOtp).toHaveBeenCalledWith('user@example.com');
    expect(result.current.step).toBe('code');
    expect(result.current.error).toBe('');
  });

  it('показывает сообщение ошибки запроса', async () => {
    mutationMocks.requestOtp.mockRejectedValueOnce(new Error('Почтовый сервис недоступен'));
    const { result } = renderHook(() => useOtpAuth());

    act(() => result.current.changeEmail('user@example.com'));
    await act(async () => result.current.sendCode());

    expect(result.current.error).toBe('Почтовый сервис недоступен');
  });

  it('оставляет в коде только восемь цифр и выполняет вход', async () => {
    const onSuccess = vi.fn();
    const { result } = renderHook(() => useOtpAuth({ onSuccess }));

    act(() => {
      result.current.changeEmail('user@example.com');
      result.current.changeCode('12a34-567890');
    });

    expect(result.current.code).toBe('12345678');
    expect(result.current.isCodeValid).toBe(true);

    await act(async () => result.current.signInByCode());

    expect(mutationMocks.verifyOtp).toHaveBeenCalledWith({
      email: 'user@example.com',
      code: '12345678',
    });
    expect(onSuccess).toHaveBeenCalledOnce();
  });

  it('очищает код и ошибку при возврате к email', () => {
    const { result } = renderHook(() => useOtpAuth());

    act(() => {
      result.current.changeCode('12345678');
      result.current.returnToEmailStep();
    });

    expect(result.current.step).toBe('email');
    expect(result.current.code).toBe('');
    expect(result.current.error).toBe('');
  });
});
