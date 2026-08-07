import type {
  TTodaySummary,
  TTodaySummaryLevelUp,
  TTodaySummaryReward,
  TTodaySummaryTask,
} from '../api/get-today-summary';

export type TTodaySummaryEvent =
  | {
      eventType: 'TASK';
      occurredAt: string;
      data: TTodaySummaryTask;
    }
  | {
      eventType: 'REWARD';
      occurredAt: string;
      data: TTodaySummaryReward;
    }
  | {
      eventType: 'LEVEL_UP';
      occurredAt: string;
      data: TTodaySummaryLevelUp;
    };

export type TTodaySummaryStats = {
  leavesCount: number;
  activitiesCount: number;
  events: TTodaySummaryEvent[];
};

export const getTodaySummaryStats = (summary: TTodaySummary): TTodaySummaryStats => {
  const taskEvents: TTodaySummaryEvent[] = summary.tasks.map((task) => ({
    eventType: 'TASK',
    occurredAt: task.completedAt,
    data: task,
  }));

  const rewardEvents: TTodaySummaryEvent[] = summary.rewards.map((reward) => ({
    eventType: 'REWARD',
    occurredAt: reward.receivedAt,
    data: reward,
  }));

  const levelUpEvents: TTodaySummaryEvent[] = summary.levelUp
    ? [
        {
          eventType: 'LEVEL_UP',
          occurredAt: summary.levelUp.occurredAt,
          data: summary.levelUp,
        },
      ]
    : [];

  const events = [...taskEvents, ...rewardEvents, ...levelUpEvents].sort(
    (a, b) => new Date(b.occurredAt).getTime() - new Date(a.occurredAt).getTime(),
  );

  return {
    leavesCount: summary.leavesEarnedToday,
    activitiesCount: events.length,
    events,
  };
};
