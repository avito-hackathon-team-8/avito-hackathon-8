import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { logout, requestOtp, verifyOtp } from './auth';

const mocks = vi.hoisted(() => ({
  apiRequest: vi.fn(),
  fetch: vi.fn(),
  removeSessionStorageValue: vi.fn(),
}));

vi.mock('@/shared/api', () => ({
  apiRequest: mocks.apiRequest,
}));

vi.mock('@/shared/lib', () => ({
  removeSessionStorageValue: mocks.removeSessionStorageValue,
  sessionStorageKeysMap: { authToken: 'auth_token' },
}));

describe('auth API', () => {
  beforeEach(() => {
    mocks.fetch.mockReset().mockResolvedValue(new Response());
    mocks.apiRequest.mockReset();
    mocks.removeSessionStorageValue.mockReset();
    vi.stubGlobal('fetch', mocks.fetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('отправляет запрос на получение одноразового кода', async () => {
    mocks.apiRequest.mockResolvedValue({ sent: true });

    await expect(requestOtp('user@example.com')).resolves.toEqual({ sent: true });
    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/app\/auth\/request-otp$/),
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: 'user@example.com' }),
      },
    );
  });

  it('отправляет email и код на подтверждение', async () => {
    const response = {
      token: 'token-1',
      record: { id: 'user-1', email: 'user@example.com', verified: true },
    };
    mocks.apiRequest.mockResolvedValue(response);

    await expect(verifyOtp('user@example.com', '12345678')).resolves.toEqual(response);
    expect(mocks.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/\/api\/app\/auth\/verify-otp$/),
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: 'user@example.com', code: '12345678' }),
      },
    );
  });

  it('удаляет токен при выходе', () => {
    logout();

    expect(mocks.removeSessionStorageValue).toHaveBeenCalledWith('auth_token');
  });
});
