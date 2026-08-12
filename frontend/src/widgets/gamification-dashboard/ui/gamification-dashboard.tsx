import { ActivityDaysCard } from '@/entities/activity-days';
import { DailyTaskCard } from '@/entities/daily-task';
import { LeaderboardCard } from '@/entities/leaderboard';
import { RewardCard } from '@/entities/reward';
import { RulesCard } from '@/entities/rules';
import { ShopThingsCard } from '@/entities/shop-things/ui/shop-things-card/shop-things-card';
import { TodaySummaryCard } from '@/entities/today-summary';

import styles from './gamification-dashboard.module.scss';

export const GamificationDashboard = () => {
  return (
    <section className={styles.gamificationDashboard}>
      <h2 className={styles.gamificationDashboard__title}>Панель информации</h2>
      <DailyTaskCard />
      <RewardCard />
      <LeaderboardCard />
      <ActivityDaysCard />
      <ShopThingsCard className={styles.gamificationDashboard__shop} />
      <TodaySummaryCard className={styles.gamificationDashboard__summary} />
      <RulesCard className={styles.gamificationDashboard__rules} />
    </section>
  );
};
