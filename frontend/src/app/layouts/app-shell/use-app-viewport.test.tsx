import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { useAppViewport } from './use-app-viewport';

const VIEWPORT_PROPERTIES = [
  '--app-viewport-bottom-gap',
  '--app-viewport-height',
  '--app-viewport-offset-top',
  '--app-visual-viewport-bottom',
];

describe('useAppViewport', () => {
  const originalInnerHeight = window.innerHeight;
  const originalInnerWidth = window.innerWidth;
  const originalScreenHeight = window.screen.height;
  const originalScreenWidth = window.screen.width;
  const originalUserAgent = window.navigator.userAgent;

  afterEach(() => {
    Object.defineProperty(window, 'innerHeight', { configurable: true, value: originalInnerHeight });
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalInnerWidth });
    Object.defineProperty(window.screen, 'height', {
      configurable: true,
      value: originalScreenHeight,
    });
    Object.defineProperty(window.screen, 'width', {
      configurable: true,
      value: originalScreenWidth,
    });
    Object.defineProperty(window.navigator, 'userAgent', {
      configurable: true,
      value: originalUserAgent,
    });

    VIEWPORT_PROPERTIES.forEach((property) => {
      document.documentElement.style.removeProperty(property);
    });
  });

  it('покрывает физическую высоту мобильного экрана под прозрачной панелью Safari', () => {
    Object.defineProperty(window, 'innerHeight', { configurable: true, value: 700 });
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 390 });
    Object.defineProperty(window.screen, 'height', { configurable: true, value: 844 });
    Object.defineProperty(window.screen, 'width', { configurable: true, value: 390 });
    Object.defineProperty(window.navigator, 'userAgent', {
      configurable: true,
      value: 'Mozilla/5.0 (iPhone; CPU iPhone OS 26_0 like Mac OS X)',
    });

    const { unmount } = renderHook(() => useAppViewport());
    const rootStyle = document.documentElement.style;

    expect(rootStyle.getPropertyValue('--app-viewport-offset-top')).toBe('0px');
    expect(rootStyle.getPropertyValue('--app-viewport-height')).toBe('844px');
    expect(rootStyle.getPropertyValue('--app-visual-viewport-bottom')).toBe('700px');
    expect(rootStyle.getPropertyValue('--app-viewport-bottom-gap')).toBe('144px');

    unmount();

    VIEWPORT_PROPERTIES.forEach((property) => {
      expect(rootStyle.getPropertyValue(property)).toBe('');
    });
  });
});
