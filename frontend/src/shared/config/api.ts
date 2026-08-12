const defaultBaseUrl = typeof window === 'undefined' ? '' : window.location.origin;

export const API_CONFIG = {
  baseUrl: import.meta.env.VITE_API_BASE_URL || defaultBaseUrl,
  basePath: import.meta.env.VITE_API_BASE_PATH || '/api',
} as const;

export const API_URL = `${API_CONFIG.baseUrl.replace(/\/$/, '')}${API_CONFIG.basePath}`;

export const resolveAssetUrl = (url: string): string => {
  const basePath = `/${API_CONFIG.basePath.replace(/^\/+|\/+$/g, '')}`;

  if (url === basePath || url.startsWith(`${basePath}/`)) {
    return `${API_CONFIG.baseUrl.replace(/\/$/, '')}${url}`;
  }

  return url;
};

export const mainQueryKey = {
  all: ['app'] as const,
};
