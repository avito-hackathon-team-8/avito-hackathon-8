import clsx from 'clsx';

import { leafIcon } from '@/shared/assets/icon';
import { Button } from '@/shared/ui/button';
import { Modal } from '@/shared/ui/modal';
import { Typography } from '@/shared/ui/typography';

import { useBuyReward } from '../../model/use-buy-reward';

import styles from './buy-reward.module.scss';

interface IBuyRewardProps {
  className?: string;
}

export const BuyReward = ({ className }: IBuyRewardProps) => {
  const {
    pet,
    reward,
    isOpen,
    isPending,
    isDisabled,
    isMVPLeavesPending,
    isRewardVisible,
    openChest,
    addMVPLeaves,
    closeModal,
  } = useBuyReward();

  return (
    <>
      <Button
        className={clsx(styles.buttonBuy, className)}
        variant="primary"
        disabled={isDisabled}
        onClick={openChest}
      >
        <Typography variant="p3-semiBold" as="span" color="inherit">
          {isPending ? 'Открываем сундук...' : 'Открыть сундук'}
        </Typography>

        <Typography
          className={clsx(styles.buttonBuy__description, {
            [styles.buttonBuy__description]: pet?.nextLevelTargetLeaves === 0,
          })}
          variant="caption"
          as="span"
          color="inherit"
        >
          {pet?.nextLevelTargetLeaves === 0 && (
            <>
              Стоимость открытия {pet.chestPrice}
              <img src={leafIcon} width={24} height={24} aria-label="Валюта листика" />
            </>
          )}

          {pet?.nextLevelTargetLeaves !== 0 && 'Разблокируется на 10 уровне'}
        </Typography>
      </Button>

      <Button
        className={styles.buttonLeaves}
        variant="primary"
        disabled={isMVPLeavesPending}
        onClick={addMVPLeaves}
      >
        <Typography variant="p3-semiBold" as="span" color="inherit">
          {isMVPLeavesPending ? 'Начисляем листья...' : 'MVP: +200 листьев'}
        </Typography>
      </Button>

      <Modal isOpen={isOpen} onClose={closeModal}>
        <div className={styles.rewardModal} aria-live="polite">
          <Typography as="h2" variant="section">
            Ваша награда
          </Typography>

          <div className={styles.rewardModal__descriptions}>
            <Typography
              as="p"
              variant="body"
              className={clsx(styles.rewardModal__description, {
                [styles.rewardModal__description_load]: !isRewardVisible,
              })}
            >
              Уговариваем {pet?.name} выдать награду…
            </Typography>

            <Typography
              as="p"
              variant="body"
              className={clsx(styles.rewardModal__description, {
                [styles.rewardModal__description_open]: isRewardVisible,
              })}
            >
              {reward?.title}
            </Typography>
          </div>

          <Button variant="primary" isFullWidth disabled={!isRewardVisible} onClick={closeModal}>
            Закрыть
          </Button>
        </div>
      </Modal>
    </>
  );
};
