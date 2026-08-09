import { type CSSProperties, useMemo } from 'react';

import { leafIcon } from '@/shared/assets/icon';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { Typography } from '@/shared/ui/typography';

import { useLevelsProfile } from '../../model/use-levels-profile';
import { usePetProfile } from '../../model/use-pet-profile';
import { LevelsPanelContent } from '../levels-panel-content/levels-panel-content';

import styles from './gamification-profile.module.scss';

export const GamificationProfile = () => {
  const { data: pet } = usePetProfile();
  const { data: levelsProfile, receiveReward } = useLevelsProfile();

  const unopenedLevel = useMemo(() => {
    if (levelsProfile) {
      return levelsProfile.levels.filter((item) => item.status === 'UNOPENED')[0];
    }
  }, [levelsProfile]);

  if (!pet) return;

  const isMaxLevel = pet.nextLevelTargetLeaves === 0;
  const percent = pet.nextLevelTargetLeaves ? (pet.leaves / pet.nextLevelTargetLeaves) * 100 : 0;

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

            {unopenedLevel && (
              <span className={styles.profile__rewardIndicator} aria-hidden="true" />
            )}

            <div className={styles.profile__info}>
              <Typography className={styles.profile__infoLevels} variant="display">
                {pet.level}/{levelsProfile?.levels.length}
              </Typography>

              <Typography
                aria-label="опыт"
                className={styles.profile__infoExperience}
                variant="p2-semiBold"
              >
                {pet.leaves} {isMaxLevel || `/${pet.nextLevelTargetLeaves}`}{' '}
                <img src={leafIcon} width={24} height={24} aria-label="Валюта из листьев" />
              </Typography>

              <div
                className={styles.profile__infoProgressBar}
                style={
                  {
                    '--progress': `${isMaxLevel ? 100 : percent}%`,
                  } as CSSProperties
                }
              >
                <div className={styles.profile__infoProgressBar_line}></div>
              </div>

              {unopenedLevel && (
                <Typography
                  className={styles.profile__infoReward}
                  variant="p4-regular"
                  color="gray500"
                >
                  Доступная награда: {unopenedLevel.reward.description}
                </Typography>
              )}
            </div>

            <button
              onClick={open}
              className={styles.profile__button}
              aria-label="Открыть список уровней с наградами"
            ></button>
          </section>
        )}
      >
        {levelsProfile && levelsProfile.levels.length > 0 && (
          <LevelsPanelContent
            levelsList={levelsProfile.levels}
            handleReceiveReward={receiveReward}
          />
        )}
      </BottomPanel>
    </section>
  );
};
