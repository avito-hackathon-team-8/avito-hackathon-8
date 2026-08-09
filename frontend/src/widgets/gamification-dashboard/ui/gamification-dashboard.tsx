import { ActivityDaysCard } from '@/entities/activity-days';
import { DailyTaskCard } from '@/entities/daily-task';
import { LeaderboardCard } from '@/entities/leaderboard';
import { RewardCard } from '@/entities/reward';
import { RulesCard } from '@/entities/rules';
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
      <TodaySummaryCard className={styles.gamificationDashboard__summary} />
      <RulesCard className={styles.gamificationDashboard__rules} />
    </section>
  );
};
