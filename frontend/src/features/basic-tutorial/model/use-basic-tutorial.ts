import { useCallback, useEffect, useRef } from 'react';

import { type Driver, driver, type DriverHook } from 'driver.js';

const BASIC_TUTORIAL_STORAGE_KEY = 'basic-tutorial-version';
const BASIC_TUTORIAL_VERSION = '1';
const AUTO_START_DELAY_MS = 2_200;
const DIALOG_RECHECK_DELAY_MS = 500;

type TUseBasicTutorialParams = {
  enabled: boolean;
};

const scrollTutorialTargetIntoView: DriverHook = (element) => {
  element?.scrollIntoView({
    behavior: 'auto',
    block: 'center',
    inline: 'nearest',
  });
};

export const useBasicTutorial = ({ enabled }: TUseBasicTutorialParams) => {
  const tutorialRef = useRef<Driver | null>(null);

  const startTutorial = useCallback(() => {
    tutorialRef.current?.destroy();

    const visualViewport = window.visualViewport;
    const refreshTutorial = () => tutorialRef.current?.refresh();

    const tutorial = driver({
      allowClose: true,
      disableActiveInteraction: true,
      doneBtnText: 'Готово',
      nextBtnText: 'Далее',
      overlayClickBehavior: 'close',
      overlayColor: '#0f0f0f',
      overlayOpacity: 0.58,
      onDoneClick: (_element, _step, { driver: tutorialDriver }) => {
        tutorialDriver.destroy();
        document.querySelector('[data-tutorial="pet"]')?.scrollIntoView({
          behavior: 'smooth',
          block: 'start',
          inline: 'nearest',
        });
      },
      onHighlightStarted: scrollTutorialTargetIntoView,
      popoverClass: 'basic-tutorial-popover',
      prevBtnText: 'Назад',
      progressText: '{{current}} из {{total}}',
      showProgress: true,
      skipMissingElement: true,
      smoothScroll: false,
      stagePadding: 8,
      stageRadius: 20,
      waitForElement: 3_000,
      steps: [
        {
          element: '[data-tutorial="pet"]',
          popover: {
            title: 'Ваш питомец',
            description:
              'Это комната питомца. Повышайте его уровень и украшайте комнату аксессуарами.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="progress"]',
          popover: {
            title: 'Уровень и листья',
            description:
              'Листья повышают уровень питомца и используются для покупок. Нажмите на блок, чтобы посмотреть награды уровней.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="pet-care"]',
          popover: {
            title: 'Забота о питомце',
            description:
              'Кормите и гладьте питомца, чтобы повышать его настроение. После использования кнопка снова станет доступна через 6 часов.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="tasks"]',
          popover: {
            title: 'Ежедневные задания',
            description:
              'Выполняйте задания на Авито и забирайте листья. Прогресс обновляется автоматически.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="rewards"]',
          popover: {
            title: 'Награды',
            description:
              'Здесь хранятся полученные бонусы. Следите за сроком действия и применяйте их вовремя.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="leaderboard"]',
          popover: {
            title: 'Лидерборд',
            description:
              'Сравнивайте свой прогресс с другими участниками. Рейтинг обновляется по мере активности игроков.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="activity"]',
          popover: {
            title: 'Дни активности',
            description:
              'Заходите каждый день и забирайте награду, чтобы продолжать серию активности.',
            side: 'bottom',
          },
        },
        {
          element: '[data-tutorial="shop"]',
          popover: {
            title: 'Магазин аксессуаров',
            description:
              'Покупайте за листья миски и лежанки. Активные предметы появятся в комнате питомца.',
            side: 'top',
          },
        },
        {
          element: '[data-tutorial="summary"]',
          popover: {
            title: 'Сводка дня',
            description:
              'Здесь собраны выполненные действия, полученные листья, награды и изменения уровня за сегодня.',
            side: 'top',
          },
        },
        {
          element: '[data-tutorial="rules"]',
          popover: {
            title: 'Правила и обучение',
            description:
              'Здесь можно подробнее изучить игровые механики и повторно запустить это обучение.',
            side: 'top',
          },
        },
      ],
      onDestroyed: () => {
        visualViewport?.removeEventListener('resize', refreshTutorial);
        visualViewport?.removeEventListener('scroll', refreshTutorial);
        localStorage.setItem(BASIC_TUTORIAL_STORAGE_KEY, BASIC_TUTORIAL_VERSION);
        tutorialRef.current = null;
      },
    });

    tutorialRef.current = tutorial;
    visualViewport?.addEventListener('resize', refreshTutorial);
    visualViewport?.addEventListener('scroll', refreshTutorial);
    tutorial.drive();
  }, []);

  useEffect(() => {
    if (!enabled || localStorage.getItem(BASIC_TUTORIAL_STORAGE_KEY) === BASIC_TUTORIAL_VERSION) {
      return;
    }

    let timeoutId: number;

    const startWhenPageIsReady = () => {
      const hasOpenDialog = document.querySelector('[data-open="true"] [role="dialog"]');

      if (hasOpenDialog) {
        timeoutId = window.setTimeout(startWhenPageIsReady, DIALOG_RECHECK_DELAY_MS);

        return;
      }

      startTutorial();
    };

    timeoutId = window.setTimeout(startWhenPageIsReady, AUTO_START_DELAY_MS);

    return () => window.clearTimeout(timeoutId);
  }, [enabled, startTutorial]);

  useEffect(() => {
    return () => tutorialRef.current?.destroy();
  }, []);

  return { startTutorial };
};
