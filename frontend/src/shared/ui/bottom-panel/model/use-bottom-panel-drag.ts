import {
  type CSSProperties,
  type MouseEvent as ReactMouseEvent,
  type RefObject,
  useEffect,
  useRef,
  useState,
} from 'react';

type UseBottomPanelDragParams = {
  panelRef: RefObject<HTMLElement | null>;
  onClose: () => void;
};

export const useBottomPanelDrag = ({ panelRef, onClose }: UseBottomPanelDragParams) => {
  const dragStartYRef = useRef(0);

  const [dragOffset, setDragOffset] = useState(0);
  const [isDragging, setIsDragging] = useState(false);

  const handleDragStart = (event: ReactMouseEvent<HTMLButtonElement>) => {
    if (event.button !== 0) {
      return;
    }

    event.preventDefault();

    dragStartYRef.current = event.clientY;
    setIsDragging(true);
  };

  useEffect(() => {
    if (!isDragging) {
      return;
    }

    const handleMouseMove = (event: MouseEvent) => {
      event.preventDefault();

      setDragOffset(Math.max(0, event.clientY - dragStartYRef.current));
    };

    const handleMouseUp = (event: MouseEvent) => {
      const panelHeight = panelRef.current?.getBoundingClientRect().height ?? 0;
      const finalDragOffset = Math.max(0, event.clientY - dragStartYRef.current);

      setIsDragging(false);
      setDragOffset(0);

      if (panelHeight > 0 && finalDragOffset >= panelHeight / 2) {
        onClose();
      }
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isDragging, onClose, panelRef]);

  const panelStyle: CSSProperties = {
    transform: dragOffset > 0 ? `translateY(${dragOffset}px)` : undefined,
    transition: isDragging ? 'none' : undefined,
  };

  return {
    handleDragStart,
    panelStyle,
  };
};
