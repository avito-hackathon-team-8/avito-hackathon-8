const sessionStorageKeys = {
  authToken: "auth_token",
} as const;

export type SessionStorageKey =
  (typeof sessionStorageKeys)[keyof typeof sessionStorageKeys];

const getStorage = () => sessionStorage;

export const getSessionStorageValue = (key: SessionStorageKey) => {
  return getStorage().getItem(key);
};

export const setSessionStorageValue = (
  key: SessionStorageKey,
  value: string,
) => {
  getStorage().setItem(key, value);
};

export const removeSessionStorageValue = (key: SessionStorageKey) => {
  getStorage().removeItem(key);
};

export const sessionStorageKeysMap = sessionStorageKeys;
