import { describe, expect, it } from 'vitest';

import { API_CONFIG, resolveAssetUrl } from './api';

describe('resolveAssetUrl', () => {
  it('добавляет backend origin к API-изображению', () => {
    expect(resolveAssetUrl('/api/v1/shop-images/bowl-fashionable.webp')).toBe(
      `${API_CONFIG.baseUrl.replace(/\/$/, '')}/api/v1/shop-images/bowl-fashionable.webp`,
    );
  });

  it.each([
    '/assets/bowl-fashionable.webp',
    'https://cdn.example.com/bowl.webp',
    'data:image/webp;base64,image',
    'blob:http://localhost:5173/image-id',
  ])('не изменяет frontend или абсолютный URL %s', (url) => {
    expect(resolveAssetUrl(url)).toBe(url);
  });
});
