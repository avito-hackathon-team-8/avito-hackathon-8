import clsx from 'clsx';

import type { User } from '@/entities/user';
import { leafIcon } from '@/shared/assets/icon';
import { Typography } from '@/shared/ui/typography';

import type { TLeaderboardUser } from '../../api/leaderboard';

import styles from './leaderboard-panel-content.module.scss';

interface ILeaderboardPanelContentProps {
  listUsers: TLeaderboardUser[];
  userPosition: number;
  user: User;
}

export const LeaderboardPanelContent = ({
  listUsers,
  userPosition,
  user,
}: ILeaderboardPanelContentProps) => {
  const hasUserTop = listUsers.some((item) => item.playerId === user?.id);

  if (!user) return;

  return (
    <ul className={styles.leaderboardPanel}>
      {listUsers.map(({ playerId, nickname, position, leaves }) => (
        <li
          className={clsx(styles.leaderboardPanel__item, {
            [styles.leaderboardPanel__item_active]: user.id === playerId,
          })}
          key={playerId}
        >
          <Typography className={styles.leaderboardPanel__text} variant="caption">
            {position}
          </Typography>
          <Typography className={styles.leaderboardPanel__text} variant="caption">
            {user.id === playerId ? 'вы' : nickname}
          </Typography>
          <span className={styles.leaderboardPanel__leaves}>
            <img src={leafIcon} width={20} height={20} aria-hidden />
            <Typography variant="caption">{leaves}</Typography>
          </span>
        </li>
      ))}

      {!hasUserTop && (
        <>
          <li className={clsx(styles.leaderboardPanel__item, styles.leaderboardPanel__item_points)}>
            <Typography variant="p3" as="span" color="gray500">
              •••
            </Typography>
          </li>
          <li className={clsx(styles.leaderboardPanel__item, styles.leaderboardPanel__item_active)}>
            <Typography className={styles.leaderboardPanel__text} variant="caption">
              {userPosition}
            </Typography>
            <Typography className={styles.leaderboardPanel__text} variant="caption">
              Вы
            </Typography>
            <span className={styles.leaderboardPanel__leaves}>
              <img src={leafIcon} width={20} height={20} aria-hidden />
              <Typography variant="caption">{user.leaderboard?.player.leaves ?? 0}</Typography>
            </span>
          </li>
        </>
      )}
    </ul>
  );
};
