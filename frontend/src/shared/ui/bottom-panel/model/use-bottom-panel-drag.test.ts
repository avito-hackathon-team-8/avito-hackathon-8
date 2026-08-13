import { act, renderHook } from '@testing-library/react';
import type { PointerEvent as ReactPointerEvent, RefObject } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useBottomPanelDrag } from './use-bottom-panel-drag';

type PointerEventOverrides = Partial<
  Pick<ReactPointerEvent<HTMLElement>, 'button' | 'clientY' | 'isPrimary' | 'pointerId'>
>;

const createPointerTarget = () =>
  ({
    setPointerCapture: vi.fn(),
    hasPointerCapture: vi.fn(() => true),
    releasePointerCapture: vi.fn(),
  }) as unknown as HTMLElement;

const createPointerEvent = (currentTarget: HTMLElement, overrides: PointerEventOverrides = {}) =>
  ({
    button: 0,
    clientY: 100,
    isPrimary: true,
    pointerId: 1,
    preventDefault: vi.fn(),
    currentTarget,
    ...overrides,
  }) as unknown as ReactPointerEvent<HTMLElement>;

const createPanelRef = (height = 400): RefObject<HTMLElement | null> => {
  const panel = document.createElement('section');
  vi.spyOn(panel, 'getBoundingClientRect').mockReturnValue({
    bottom: height,
    height,
    left: 0,
    right: 320,
    top: 0,
    width: 320,
    x: 0,
    y: 0,
    toJSON: () => ({}),
  });

  return { current: panel };
};

describe('useBottomPanelDrag', () => {
  const onClose = vi.fn();

  beforeEach(() => {
    onClose.mockReset();
  });

  it('начинает основной drag и применяет смещение вниз', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(), onClose }),
    );
    const startEvent = createPointerEvent(target, { clientY: 100 });
    const moveEvent = createPointerEvent(target, { clientY: 250 });

    act(() => result.current.handleDragStart(startEvent));
    expect(startEvent.preventDefault).toHaveBeenCalledOnce();
    expect(target.setPointerCapture).toHaveBeenCalledWith(1);

    act(() => result.current.handleDragMove(moveEvent));
    expect(moveEvent.preventDefault).toHaveBeenCalledOnce();
    expect(result.current.panelStyle).toEqual({
      transform: 'translateY(150px)',
      transition: 'none',
    });
  });

  it('игнорирует неосновной указатель и события другого pointerId', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(), onClose }),
    );
    const invalidStart = createPointerEvent(target, { isPrimary: false });

    act(() => result.current.handleDragStart(invalidStart));
    expect(invalidStart.preventDefault).not.toHaveBeenCalled();

    act(() => result.current.handleDragStart(createPointerEvent(target)));
    act(() =>
      result.current.handleDragMove(createPointerEvent(target, { pointerId: 2, clientY: 250 })),
    );
    expect(result.current.panelStyle.transform).toBeUndefined();
  });

  it('не позволяет перетаскивать панель вверх', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(), onClose }),
    );

    act(() => result.current.handleDragStart(createPointerEvent(target, { clientY: 100 })));
    act(() => result.current.handleDragMove(createPointerEvent(target, { clientY: 50 })));

    expect(result.current.panelStyle.transform).toBeUndefined();
    expect(result.current.panelStyle.transition).toBe('none');
  });

  it('возвращает панель на место, если порог закрытия не достигнут', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(400), onClose }),
    );

    act(() => result.current.handleDragStart(createPointerEvent(target, { clientY: 100 })));
    act(() => result.current.handleDragEnd(createPointerEvent(target, { clientY: 299 })));

    expect(target.releasePointerCapture).toHaveBeenCalledWith(1);
    expect(result.current.panelStyle).toEqual({ transform: undefined, transition: undefined });
    expect(onClose).not.toHaveBeenCalled();
  });

  it('закрывает панель при смещении минимум на половину высоты', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(400), onClose }),
    );

    act(() => result.current.handleDragStart(createPointerEvent(target, { clientY: 100 })));
    act(() => result.current.handleDragEnd(createPointerEvent(target, { clientY: 300 })));

    expect(onClose).toHaveBeenCalledOnce();
  });

  it('сбрасывает drag-состояние при отмене указателя', () => {
    const target = createPointerTarget();
    const { result } = renderHook(() =>
      useBottomPanelDrag({ panelRef: createPanelRef(), onClose }),
    );

    act(() => result.current.handleDragStart(createPointerEvent(target)));
    act(() => result.current.handleDragMove(createPointerEvent(target, { clientY: 220 })));
    act(() => result.current.handleDragCancel(createPointerEvent(target)));

    expect(result.current.panelStyle).toEqual({ transform: undefined, transition: undefined });
    expect(onClose).not.toHaveBeenCalled();
  });
});
