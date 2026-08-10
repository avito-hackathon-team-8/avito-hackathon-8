import { useEffect } from 'react';

let activeLocksCount = 0;
let previousHtmlOverflow = '';
let previousBodyOverflow = '';

const lockPageScroll = () => {
  if (activeLocksCount === 0) {
    previousHtmlOverflow = document.documentElement.style.overflow;
    previousBodyOverflow = document.body.style.overflow;

    document.documentElement.style.overflow = 'hidden';
    document.body.style.overflow = 'hidden';
  }

  activeLocksCount += 1;
};

const unlockPageScroll = () => {
  activeLocksCount = Math.max(0, activeLocksCount - 1);

  if (activeLocksCount === 0) {
    document.documentElement.style.overflow = previousHtmlOverflow;
    document.body.style.overflow = previousBodyOverflow;
  }
};

export const usePageScrollLock = (isLocked: boolean) => {
  useEffect(() => {
    if (!isLocked) {
      return;
    }

    lockPageScroll();

    return unlockPageScroll;
  }, [isLocked]);
};
