import { apiRequest } from '@/shared/api/api-request';
import { removeSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib/session-storage';

const AUTH_API = '/api/app/auth';

export type User = {
  id: string;
  email: string;
  verified: boolean;
  position?: number;
};

export type AuthResponse = {
  token: string;
  record: User;
};

export const requestOtp = (email: string): Promise<{ sent: boolean }> =>
  apiRequest(
    fetch(`${AUTH_API}/request-otp`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    }),
  );

export const verifyOtp = async (email: string, code: string): Promise<AuthResponse> =>
  apiRequest(
    fetch(`${AUTH_API}/verify-otp`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, code }),
    }),
  );

export const getCurrentUser = async (token: string | null): Promise<User> => {
  return await apiRequest(
    fetch(`${AUTH_API}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    }),
    'Ошибка получения данных getCurrentUser',
    () => removeSessionStorageValue(sessionStorageKeysMap.authToken),
  );
};

export const logout = () => {
  removeSessionStorageValue(sessionStorageKeysMap.authToken);
};
