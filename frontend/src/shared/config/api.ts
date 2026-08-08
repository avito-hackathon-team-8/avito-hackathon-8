export const API_CONFIG = {
  baseUrl: import.meta.env.VITE_API_BASE_URL,
  basePath: import.meta.env.VITE_API_BASE_PATH,
} as const;

export const API_URL = `${API_CONFIG.baseUrl.replace(/\/$/, '')}${API_CONFIG.basePath}`;

export const mainQueryKey = {
  all: ['app'] as const,
};
