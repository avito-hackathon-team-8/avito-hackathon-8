import { useEffect } from 'react';

let activeLocksCount = 0;
let previousHtmlOverflow = '';
let previousBodyOverflow = '';
let previousBodyPosition = '';
let previousBodyTop = '';
let previousBodyRight = '';
let previousBodyLeft = '';
let previousBodyWidth = '';
let lockedScrollY = 0;

const lockPageScroll = () => {
  if (activeLocksCount === 0) {
    previousHtmlOverflow = document.documentElement.style.overflow;
    previousBodyOverflow = document.body.style.overflow;
    previousBodyPosition = document.body.style.position;
    previousBodyTop = document.body.style.top;
    previousBodyRight = document.body.style.right;
    previousBodyLeft = document.body.style.left;
    previousBodyWidth = document.body.style.width;
    lockedScrollY = window.scrollY;

    document.documentElement.style.overflow = 'hidden';
    document.body.style.overflow = 'hidden';
    document.body.style.position = 'fixed';
    document.body.style.top = `-${lockedScrollY}px`;
    document.body.style.right = '0';
    document.body.style.left = '0';
    document.body.style.width = '100%';
  }

  activeLocksCount += 1;
};

const unlockPageScroll = () => {
  activeLocksCount = Math.max(0, activeLocksCount - 1);

  if (activeLocksCount === 0) {
    document.documentElement.style.overflow = previousHtmlOverflow;
    document.body.style.overflow = previousBodyOverflow;
    document.body.style.position = previousBodyPosition;
    document.body.style.top = previousBodyTop;
    document.body.style.right = previousBodyRight;
    document.body.style.left = previousBodyLeft;
    document.body.style.width = previousBodyWidth;

    if (lockedScrollY > 0) {
      window.scrollTo(0, lockedScrollY);
    }
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
