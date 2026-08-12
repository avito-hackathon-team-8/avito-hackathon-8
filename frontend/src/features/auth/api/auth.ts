import type { User } from '@/entities/user';
import { apiRequest } from '@/shared/api';
import { API_URL } from '@/shared/config';
import { removeSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib';

const AUTH_API = `${API_URL}/app/auth`;

type TAuthUserRecord = Pick<User, 'id' | 'email' | 'verified'>;

export type AuthResponse = {
  token: string;
  record: TAuthUserRecord;
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

export const logout = () => {
  removeSessionStorageValue(sessionStorageKeysMap.authToken);
};
