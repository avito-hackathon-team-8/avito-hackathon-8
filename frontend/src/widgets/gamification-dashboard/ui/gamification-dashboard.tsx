import { ActivityDaysCard } from '@/entities/activity-days';
import { DailyTaskCard } from '@/entities/daily-task';
import { LeaderboardCard } from '@/entities/leaderboard';
import { RewardCard } from '@/entities/reward';
import { RulesCard } from '@/entities/rules';
import { ShopThingsCard } from '@/entities/shop-things/ui/shop-things-card/shop-things-card';
import { TodaySummaryCard } from '@/entities/today-summary';

import styles from './gamification-dashboard.module.scss';

type TGamificationDashboardProps = {
  onStartTutorial?: () => void;
};

export const GamificationDashboard = ({ onStartTutorial }: TGamificationDashboardProps) => {
  return (
    <section className={styles.gamificationDashboard}>
      <h2 className={styles.gamificationDashboard__title}>Панель информации</h2>
      <div className={styles.gamificationDashboard__tutorialTarget} data-tutorial="tasks">
        <DailyTaskCard />
      </div>
      <div className={styles.gamificationDashboard__tutorialTarget} data-tutorial="rewards">
        <RewardCard />
      </div>
      <div className={styles.gamificationDashboard__tutorialTarget} data-tutorial="leaderboard">
        <LeaderboardCard />
      </div>
      <div className={styles.gamificationDashboard__tutorialTarget} data-tutorial="activity">
        <ActivityDaysCard />
      </div>
      <div
        className={`${styles.gamificationDashboard__tutorialTarget} ${styles.gamificationDashboard__shop}`}
        data-tutorial="shop"
      >
        <ShopThingsCard />
      </div>
      <div
        className={`${styles.gamificationDashboard__tutorialTarget} ${styles.gamificationDashboard__summary}`}
        data-tutorial="summary"
      >
        <TodaySummaryCard />
      </div>
      <div
        className={`${styles.gamificationDashboard__tutorialTarget} ${styles.gamificationDashboard__rules}`}
        data-tutorial="rules"
      >
        <RulesCard onStartTutorial={onStartTutorial} />
      </div>
    </section>
  );
};
