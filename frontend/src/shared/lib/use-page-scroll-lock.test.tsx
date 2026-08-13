import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { usePageScrollLock } from './use-page-scroll-lock';

describe('usePageScrollLock', () => {
  afterEach(() => {
    document.documentElement.removeAttribute('style');
    document.body.removeAttribute('style');
    vi.restoreAllMocks();
  });

  it('блокирует прокрутку без смещения body и полностью восстанавливает стили', () => {
    const { rerender } = renderHook(({ isLocked }) => usePageScrollLock(isLocked), {
      initialProps: { isLocked: true },
    });

    expect(document.documentElement).toHaveStyle({ overflow: 'hidden' });
    expect(document.body).toHaveStyle({ overflow: 'hidden' });
    expect(document.body.style.position).toBe('');

    rerender({ isLocked: false });

    expect(document.documentElement.style.overflow).toBe('');
    expect(document.body.style.cssText).toBe('');
  });
});
