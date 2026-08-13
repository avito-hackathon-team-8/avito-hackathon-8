import { useLayoutEffect } from 'react';

const CSS_PROPERTIES = {
  bottomGap: '--app-viewport-bottom-gap',
  height: '--app-viewport-height',
  offsetTop: '--app-viewport-offset-top',
  visualBottom: '--app-visual-viewport-bottom',
} as const;

const getOrientedScreenHeight = () => {
  const isLandscape = window.innerWidth > window.innerHeight;
  const screenSizes = [window.screen.width, window.screen.height];

  return isLandscape ? Math.min(...screenSizes) : Math.max(...screenSizes);
};

const isIPhone = () => /iPhone|iPod/.test(window.navigator.userAgent);

export const useAppViewport = () => {
  useLayoutEffect(() => {
    const rootStyle = document.documentElement.style;
    const previousValues = Object.values(CSS_PROPERTIES).map((property) => [
      property,
      rootStyle.getPropertyValue(property),
    ]);

    const updateViewport = () => {
      const visualViewport = window.visualViewport;
      const offsetTop = Math.max(0, visualViewport?.offsetTop ?? 0);
      const visualHeight = visualViewport?.height ?? window.innerHeight;
      const visualBottom = offsetTop + visualHeight;
      const screenBottom = window.innerWidth < 600 && isIPhone() ? getOrientedScreenHeight() : 0;
      const viewportBottom = Math.max(
        visualBottom,
        window.innerHeight,
        screenBottom,
      );

      rootStyle.setProperty(CSS_PROPERTIES.offsetTop, `${offsetTop}px`);
      rootStyle.setProperty(CSS_PROPERTIES.height, `${viewportBottom - offsetTop}px`);
      rootStyle.setProperty(CSS_PROPERTIES.visualBottom, `${visualBottom}px`);
      rootStyle.setProperty(CSS_PROPERTIES.bottomGap, `${viewportBottom - visualBottom}px`);
    };

    updateViewport();

    window.addEventListener('orientationchange', updateViewport);
    window.addEventListener('resize', updateViewport);
    window.visualViewport?.addEventListener('resize', updateViewport);
    window.visualViewport?.addEventListener('scroll', updateViewport);

    return () => {
      window.removeEventListener('orientationchange', updateViewport);
      window.removeEventListener('resize', updateViewport);
      window.visualViewport?.removeEventListener('resize', updateViewport);
      window.visualViewport?.removeEventListener('scroll', updateViewport);

      previousValues.forEach(([property, value]) => {
        if (value) {
          rootStyle.setProperty(property, value);
        } else {
          rootStyle.removeProperty(property);
        }
      });
    };
  }, []);
};
