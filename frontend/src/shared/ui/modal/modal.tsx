import { type MouseEvent, type PropsWithChildren, useEffect } from 'react';

import { createPortal } from 'react-dom';

import styles from './modal.module.scss';

interface IModalProps extends PropsWithChildren {
  isOpen: boolean;
  onClose: () => void;
  className?: string;
}

export const Modal = ({ isOpen, onClose, className, children }: IModalProps) => {
  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    const previousOverflow = document.body.style.overflow;

    document.body.style.overflow = 'hidden';
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) {
    return null;
  }

  const handleOverlayClick = (event: MouseEvent<HTMLDivElement>) => {
    if (event.target === event.currentTarget) {
      onClose();
    }
  };

  return createPortal(
    <div className={styles.overlay} onMouseDown={handleOverlayClick} role="presentation">
      <div className={`${styles.modal} ${className ?? ''}`} role="dialog" aria-modal="true">
        {children}
      </div>
    </div>,
    document.body,
  );
};
