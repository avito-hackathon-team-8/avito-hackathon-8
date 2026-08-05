import {
  removeSessionStorageValue,
  sessionStorageKeysMap,
} from "@/shared/lib/session-storage";

const AUTH_API = "/api/app/auth";

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

type APIError = {
  message?: string;
};

const getErrorMessage = async (response: Response) => {
  const fallback = "Something went wrong. Please try again.";

  try {
    const error = (await response.json()) as APIError;

    return error.message || fallback;
  } catch {
    return fallback;
  }
};

const request = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init);

  if (!response.ok) {
    throw new Error(await getErrorMessage(response));
  }

  return response.json() as Promise<T>;
};

export const requestOtp = (email: string) =>
  request<{ sent: boolean }>(`${AUTH_API}/request-otp`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });

export const verifyOtp = async (email: string, code: string) => {
  return request<AuthResponse>(`${AUTH_API}/verify-otp`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, code }),
  });
};

export const getCurrentUser = async (token: string | null) => {
  if (!token) {
    return null;
  }

  try {
    return await request<User>(`${AUTH_API}/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch {
    removeSessionStorageValue(sessionStorageKeysMap.authToken);
    return null;
  }
};

export const logout = () => {
  removeSessionStorageValue(sessionStorageKeysMap.authToken);
};
