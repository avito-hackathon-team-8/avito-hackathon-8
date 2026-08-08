import { type MouseEvent, type PropsWithChildren, useEffect, useState } from 'react';

import { createPortal } from 'react-dom';

import styles from './modal.module.scss';

interface IModalProps extends PropsWithChildren {
  isOpen: boolean;
  onClose: () => void;
  className?: string;
}

export const Modal = ({ isOpen, onClose, className, children }: IModalProps) => {
  const [portalRoot] = useState(() => document.getElementById('app-modal-root'));

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

  const handleOverlayClick = (event: MouseEvent<HTMLDivElement>) => {
    if (event.target === event.currentTarget) {
      onClose();
    }
  };

  if (!isOpen || !portalRoot) {
    return null;
  }

  return createPortal(
    <div className={styles.overlay}>
      <div
        className={`${styles.modal} ${className ?? ''}`}
        role="dialog"
        aria-modal="true"
        onClick={handleOverlayClick}
      >
        {children}
      </div>
    </div>,
    portalRoot,
  );
};
