import type { CSSProperties } from 'react';

import { BottomPanel } from '@/shared/ui/bottom-panel';
import { Typography } from '@/shared/ui/typography';

import { LevelsPanelContent } from '../levels-panel-content/levels-panel-content';

import styles from './gamification-profile.module.scss';

export const GamificationProfile = () => {
  return (
    <section className={styles.profile}>
      <h2 className={styles.profile__title}>Информация о питомце</h2>

      <BottomPanel
        title="Список уровней"
        renderTrigger={(open) => (
          <section className={styles.profile__content}>
            <Typography variant="caption" as="h3" color="gray500">
              Уровень питомца
            </Typography>

            <div className={styles.profile__info}>
              <Typography className={styles.profile__infoLevels} variant="display">
                3/10
              </Typography>

              <Typography
                aria-label="опыт"
                className={styles.profile__infoExperience}
                variant="p2-semiBold"
              >
                380 / 560 листьев
              </Typography>

              <div
                className={styles.profile__infoProgressBar}
                style={{ '--progress': '50%' } as CSSProperties}
              >
                <div className={styles.profile__infoProgressBar_line}></div>
              </div>

              <Typography
                className={styles.profile__infoReward}
                variant="p4-regular"
                color="gray500"
              >
                Следующая награда: 100 бонусов
              </Typography>
            </div>

            <button
              onClick={open}
              className={styles.profile__button}
              aria-label="Открыть список уровней с наградами"
            ></button>
          </section>
        )}
      >
        <LevelsPanelContent />
      </BottomPanel>
    </section>
  );
};
