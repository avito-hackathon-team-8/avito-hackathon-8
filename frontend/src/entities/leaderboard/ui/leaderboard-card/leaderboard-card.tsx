import { useCurrentUser } from '@/entities/user';
import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useLeaderboard } from '../../model/use-leaderboard';
import pedestalIcon from '../assets/pedestal-icon.svg';
import { LeaderboardPanelContent } from '../leaderboard-panel-content/leaderboard-panel-content';

const TITLE_CARD = 'Лидерборд';

export const LeaderboardCard = () => {
  const { data, refetch, isPending } = useLeaderboard();
  const { data: user } = useCurrentUser();

  const userPosition = user?.leaderboard?.player.position ?? 0;

  const handleClick = () => {
    if (!data && !isPending) {
      refetch();
    }
  };

  return (
    <BottomPanel
      title={TITLE_CARD}
      description="Топ-10 платформы. Обновляется раз в день"
      disabled={!data || data.items.length === 0}
      onClick={handleClick}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={`Ваше место: ${userPosition}`}
          imageProps={{
            src: pedestalIcon,
            alt: 'Пьедестал',
            width: 96,
            height: 65,
          }}
          wrapperProps={{ onClick: open }}
        />
      )}
    >
      {data && user && (
        <LeaderboardPanelContent listUsers={data.items} user={user} userPosition={userPosition} />
      )}
    </BottomPanel>
  );
};
