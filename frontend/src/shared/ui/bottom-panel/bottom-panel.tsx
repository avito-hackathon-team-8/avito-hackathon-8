import { type MouseEvent, type ReactNode, useEffect, useId, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { Typography } from '../typography';

import styles from './bottom-panel.module.scss';

const FOCUSABLE_ELEMENTS_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

type BottomPanelProps = {
  title: ReactNode;
  description?: string;
  children: ReactNode;
  renderTrigger: (open: () => void) => ReactNode;
  closeOnBackdrop?: boolean;
  disabled?: boolean;
  onClick?: () => void;
};

export const BottomPanel = ({
  title,
  description,
  children,
  renderTrigger,
  closeOnBackdrop = true,
  disabled = false,
  onClick,
}: BottomPanelProps) => {
  const [portalRoot, setPortalRoot] = useState<HTMLElement | null>(null);

  const [isOpen, setIsOpen] = useState(false);

  const panelRef = useRef<HTMLElement>(null);
  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null);

  const titleId = useId();

  const handleOpen = () => {
    if (onClick) {
      onClick();
    }

    if (disabled) return;

    if (portalRoot) {
      setIsOpen(true);

      return;
    }

    const root = document.getElementById('app-overlay-root');

    if (!root) {
      console.error('BottomPanel: не найден элемент #app-overlay-root');

      return;
    }

    setPortalRoot(root);

    setIsOpen(true);
  };

  const handleClose = () => {
    setIsOpen(false);
  };

  const handleOverlayClick = (event: MouseEvent<HTMLDivElement>) => {
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
        setIsOpen(false);

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

      {portalRoot
        ? createPortal(
            <div
              className={styles.overlay}
              data-open={isOpen}
              aria-hidden={!isOpen}
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
              >
                <div className={styles.panel__handle} aria-hidden="true" />

                <header className={styles.panel__header}>
                  <Typography
                    className={styles.panel__title}
                    id={titleId}
                    variant="section"
                    as="h2"
                  >
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
                  >
                    ×
                  </button>
                </header>

                <div className={styles.panel__content}>{children}</div>
              </section>
            </div>,
            portalRoot,
          )
        : null}
    </>
  );
};

export type { BottomPanelProps };
