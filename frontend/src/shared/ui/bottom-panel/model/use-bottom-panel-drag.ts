import {
  type CSSProperties,
  type PointerEvent as ReactPointerEvent,
  type RefObject,
  useRef,
  useState,
} from 'react';

type UseBottomPanelDragParams = {
  panelRef: RefObject<HTMLElement | null>;
  onClose: () => void;
};

export const useBottomPanelDrag = ({ panelRef, onClose }: UseBottomPanelDragParams) => {
  const dragStartYRef = useRef(0);
  const activePointerIdRef = useRef<number | null>(null);

  const [dragOffset, setDragOffset] = useState(0);
  const [isDragging, setIsDragging] = useState(false);

  const handleDragStart = (event: ReactPointerEvent<HTMLElement>) => {
    if (!event.isPrimary || event.button !== 0) {
      return;
    }

    event.preventDefault();
    event.currentTarget.setPointerCapture(event.pointerId);

    activePointerIdRef.current = event.pointerId;
    dragStartYRef.current = event.clientY;
    setIsDragging(true);
  };

  const handleDragMove = (event: ReactPointerEvent<HTMLElement>) => {
    if (activePointerIdRef.current !== event.pointerId) {
      return;
    }

    event.preventDefault();

    setDragOffset(Math.max(0, event.clientY - dragStartYRef.current));
  };

  const handleDragEnd = (event: ReactPointerEvent<HTMLElement>) => {
    if (activePointerIdRef.current !== event.pointerId) {
      return;
    }

    const panelHeight = panelRef.current?.getBoundingClientRect().height ?? 0;
    const finalDragOffset = Math.max(0, event.clientY - dragStartYRef.current);

    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }

    activePointerIdRef.current = null;
    setIsDragging(false);
    setDragOffset(0);

    if (panelHeight > 0 && finalDragOffset >= panelHeight / 2) {
      onClose();
    }
  };

  const handleDragCancel = (event: ReactPointerEvent<HTMLElement>) => {
    if (activePointerIdRef.current !== event.pointerId) {
      return;
    }

    activePointerIdRef.current = null;
    setIsDragging(false);
    setDragOffset(0);
  };

  const panelStyle: CSSProperties = {
    transform: dragOffset > 0 ? `translateY(${dragOffset}px)` : undefined,
    transition: isDragging ? 'none' : undefined,
  };

  return {
    handleDragCancel,
    handleDragEnd,
    handleDragMove,
    handleDragStart,
    panelStyle,
  };
};
