import { type MouseEvent, type ReactNode, useEffect, useId, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { usePageScrollLock } from '@/shared/lib';

import { Typography } from '../typography';

import { useBottomPanelDrag } from './model/use-bottom-panel-drag';

import styles from './bottom-panel.module.scss';

const FOCUSABLE_ELEMENTS_SELECTOR = [
  'a[href]',
  'button:not([disabled]):not([tabindex="-1"])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

type BottomPanelProps = {
  title: ReactNode;
  description?: string;
  children?: ReactNode;
  renderContent?: (close: () => void) => ReactNode;
  renderTrigger: (open: () => void) => ReactNode;
  closeOnBackdrop?: boolean;
  disabled?: boolean;
  onClick?: () => void;
  isStartOpen?: boolean;
};

export const BottomPanel = ({
  title,
  description,
  children,
  renderContent,
  renderTrigger,
  closeOnBackdrop = true,
  disabled = false,
  onClick,
  isStartOpen = false,
}: BottomPanelProps) => {
  const [portalRoot] = useState<HTMLElement | null>(() =>
    document.getElementById('app-overlay-root'),
  );

  const [isOpen, setIsOpen] = useState(isStartOpen);

  usePageScrollLock(isOpen);

  const panelRef = useRef<HTMLElement | null>(null);
  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null);

  const titleId = useId();

  const { handleDragCancel, handleDragEnd, handleDragMove, handleDragStart, panelStyle } =
    useBottomPanelDrag({
      panelRef,
      onClose: handleClose,
    });

  function handleOpen() {
    onClick?.();

    if (disabled) {
      return;
    }

    if (!portalRoot) {
      console.error('BottomPanel: не найден элемент #app-overlay-root');

      return;
    }

    setIsOpen(true);
  }

  function handleClose() {
    setIsOpen(false);
  }

  const handleOverlayClick = (event: MouseEvent) => {
    const isBackdropClick = event.target === event.currentTarget;

    if (closeOnBackdrop && isBackdropClick) {
      handleClose();
    }
  };

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const panel = panelRef.current;

    if (!panel) {
      return;
    }

    previouslyFocusedElementRef.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;

    const appContent =
      portalRoot?.parentElement?.querySelector<HTMLElement>('[data-app-content]') ?? null;

    if (appContent) {
      appContent.inert = true;
    }

    const getFocusableElements = () => {
      return Array.from(panel.querySelectorAll<HTMLElement>(FOCUSABLE_ELEMENTS_SELECTOR)).filter(
        (element) => {
          const isVisible = element.getClientRects().length > 0;
          const isAriaHidden = element.getAttribute('aria-hidden') === 'true';

          return isVisible && !isAriaHidden;
        },
      );
    };

    const focusFirstElement = () => {
      const focusableElements = getFocusableElements();

      const firstElement = focusableElements.at(0) ?? panel;

      firstElement.focus({
        preventScroll: true,
      });
    };

    focusFirstElement();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();

        handleClose();

        return;
      }

      if (event.key !== 'Tab') {
        return;
      }

      const focusableElements = getFocusableElements();

      if (focusableElements.length === 0) {
        event.preventDefault();

        panel.focus({
          preventScroll: true,
        });

        return;
      }

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      const activeElement = document.activeElement;

      const focusIsOutside = activeElement === null || !panel.contains(activeElement);

      if (event.shiftKey && (activeElement === firstElement || focusIsOutside)) {
        event.preventDefault();

        lastElement.focus({
          preventScroll: true,
        });

        return;
      }

      if (!event.shiftKey && (activeElement === lastElement || focusIsOutside)) {
        event.preventDefault();

        firstElement.focus({
          preventScroll: true,
        });
      }
    };

    const handleFocusIn = (event: FocusEvent) => {
      if (event.target instanceof Node && !panel.contains(event.target)) {
        focusFirstElement();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    document.addEventListener('focusin', handleFocusIn);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('focusin', handleFocusIn);

      if (appContent) {
        appContent.inert = false;
      }

      previouslyFocusedElementRef.current?.focus({
        preventScroll: true,
      });

      previouslyFocusedElementRef.current = null;
    };
  }, [isOpen, portalRoot]);

  return (
    <>
      {renderTrigger(handleOpen)}

      {portalRoot &&
        createPortal(
          <div
            className={styles.overlay}
            data-open={isOpen}
            inert={!isOpen}
            onClick={handleOverlayClick}
          >
            <section
              ref={panelRef}
              className={styles.panel}
              role="dialog"
              aria-modal="true"
              aria-labelledby={titleId}
              tabIndex={-1}
              style={panelStyle}
            >
              <header
                className={styles.panel__header}
                onPointerCancel={handleDragCancel}
                onPointerDown={handleDragStart}
                onPointerMove={handleDragMove}
                onPointerUp={handleDragEnd}
              >
                <Typography className={styles.panel__title} id={titleId} variant="section" as="h2">
                  {title}
                </Typography>

                {description && (
                  <Typography
                    className={styles.panel__description}
                    color="gray500"
                    variant="p4-bold"
                  >
                    {description}
                  </Typography>
                )}

                <button
                  type="button"
                  className={styles.panel__close}
                  aria-label="Закрыть панель"
                  onClick={handleClose}
                  onPointerDown={(event) => event.stopPropagation()}
                >
                  <Typography
                    className={styles.panel__closeText}
                    as="span"
                    color="black"
                    variant="display-normal"
                  >
                    ×
                  </Typography>
                </button>
              </header>

              <div className={styles.panel__content}>
                {renderContent ? renderContent(handleClose) : children}
              </div>
            </section>
          </div>,
          portalRoot,
        )}
    </>
  );
};

export type { BottomPanelProps };
