import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { usePageScrollLock } from './use-page-scroll-lock';

describe('usePageScrollLock', () => {
  afterEach(() => {
    document.documentElement.removeAttribute('style');
    document.body.removeAttribute('style');
    vi.restoreAllMocks();
  });

  it('фиксирует страницу на текущей позиции и полностью восстанавливает её', () => {
    Object.defineProperty(window, 'scrollY', { configurable: true, value: 240 });
    const scrollTo = vi.spyOn(window, 'scrollTo').mockImplementation(() => undefined);

    const { rerender } = renderHook(({ isLocked }) => usePageScrollLock(isLocked), {
      initialProps: { isLocked: true },
    });

    expect(document.documentElement).toHaveStyle({ overflow: 'hidden' });
    expect(document.body).toHaveStyle({
      left: '0px',
      overflow: 'hidden',
      position: 'fixed',
      right: '0px',
      top: '-240px',
      width: '100%',
    });

    rerender({ isLocked: false });

    expect(document.documentElement.style.overflow).toBe('');
    expect(document.body.style.cssText).toBe('');
    expect(scrollTo).toHaveBeenCalledWith(0, 240);
  });
});
