import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useBasicTutorial } from './use-basic-tutorial';

const mocks = vi.hoisted(() => ({
  destroy: vi.fn(),
  drive: vi.fn(),
  driver: vi.fn(),
}));

vi.mock('driver.js', () => ({
  driver: mocks.driver,
}));

describe('useBasicTutorial', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    localStorage.clear();
    mocks.destroy.mockReset();
    mocks.drive.mockReset();
    mocks.driver.mockReset().mockReturnValue({
      destroy: mocks.destroy,
      drive: mocks.drive,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    localStorage.clear();
  });

  it('автоматически запускает базовое обучение один раз', () => {
    renderHook(() => useBasicTutorial({ enabled: true }));

    act(() => vi.advanceTimersByTime(2_199));
    expect(mocks.driver).not.toHaveBeenCalled();

    act(() => vi.advanceTimersByTime(1));
    expect(mocks.driver).toHaveBeenCalledOnce();
    expect(mocks.drive).toHaveBeenCalledOnce();

    const config = mocks.driver.mock.calls[0][0];
    expect(config.steps).toHaveLength(10);
    expect(config.steps[0].element).toBe('[data-tutorial="pet"]');
    expect(config.steps[2].element).toBe('[data-tutorial="pet-care"]');
    expect(config.steps[4].element).toBe('[data-tutorial="rewards"]');
    expect(config.steps[5].element).toBe('[data-tutorial="leaderboard"]');
    expect(config.steps[6].element).toBe('[data-tutorial="activity"]');
    expect(config.steps[8].element).toBe('[data-tutorial="summary"]');
    expect(config.steps[9].element).toBe('[data-tutorial="rules"]');

    act(() => config.onDestroyed());
    expect(localStorage.getItem('basic-tutorial-version')).toBe('1');
  });

  it('не запускается автоматически после прохождения текущей версии', () => {
    localStorage.setItem('basic-tutorial-version', '1');

    renderHook(() => useBasicTutorial({ enabled: true }));
    act(() => vi.runAllTimers());

    expect(mocks.driver).not.toHaveBeenCalled();
  });

  it('ждёт закрытия нижней панели перед автозапуском', () => {
    const overlay = document.createElement('div');
    const dialog = document.createElement('section');
    overlay.dataset.open = 'true';
    dialog.setAttribute('role', 'dialog');
    overlay.append(dialog);
    document.body.append(overlay);

    renderHook(() => useBasicTutorial({ enabled: true }));
    act(() => vi.advanceTimersByTime(2_200));
    expect(mocks.driver).not.toHaveBeenCalled();

    overlay.dataset.open = 'false';
    act(() => vi.advanceTimersByTime(500));
    expect(mocks.driver).toHaveBeenCalledOnce();

    overlay.remove();
  });

  it('позволяет повторно запустить обучение вручную', () => {
    localStorage.setItem('basic-tutorial-version', '1');
    const { result } = renderHook(() => useBasicTutorial({ enabled: true }));

    act(() => result.current.startTutorial());

    expect(mocks.driver).toHaveBeenCalledOnce();
    expect(mocks.drive).toHaveBeenCalledOnce();
  });

  it('прокручивает обрезанный target в центр внутреннего контейнера', () => {
    const target = document.createElement('div');
    const scrollIntoView = vi.fn();
    target.scrollIntoView = scrollIntoView;
    const { result } = renderHook(() => useBasicTutorial({ enabled: false }));

    act(() => result.current.startTutorial());

    const config = mocks.driver.mock.calls[0][0];
    act(() => config.onHighlightStarted(target));

    expect(config.smoothScroll).toBe(false);
    expect(scrollIntoView).toHaveBeenCalledWith({
      behavior: 'auto',
      block: 'center',
      inline: 'nearest',
    });
  });

  it('после завершения закрывает обучение и прокручивает приложение к началу', () => {
    const firstTarget = document.createElement('div');
    const scrollIntoView = vi.fn();
    firstTarget.dataset.tutorial = 'pet';
    firstTarget.scrollIntoView = scrollIntoView;
    document.body.append(firstTarget);
    const { result } = renderHook(() => useBasicTutorial({ enabled: false }));

    act(() => result.current.startTutorial());

    const config = mocks.driver.mock.calls[0][0];
    act(() =>
      config.onDoneClick(undefined, config.steps[9], {
        driver: { destroy: mocks.destroy },
      }),
    );

    expect(mocks.destroy).toHaveBeenCalledOnce();
    expect(scrollIntoView).toHaveBeenCalledWith({
      behavior: 'smooth',
      block: 'start',
      inline: 'nearest',
    });

    firstTarget.remove();
  });
});
