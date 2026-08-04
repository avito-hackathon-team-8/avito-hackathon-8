import activityCalendar from '@/shared/assets/pet-dashboard/activity-calendar.svg';
import awardCup from '@/shared/assets/pet-dashboard/award-cup.svg';
import dailySummary from '@/shared/assets/pet-dashboard/daily-summary.svg';
import leaderboard from '@/shared/assets/pet-dashboard/leaderboard.svg';
import petScene from '@/shared/assets/pet-dashboard/pet-scene.svg';
import ruleBook from '@/shared/assets/pet-dashboard/rule-book.svg';
import tasksBoard from '@/shared/assets/pet-dashboard/tasks-board.svg';
import { Header } from '@/widgets/header';

import styles from './main.page.module.scss';

const cards = [
  {
    title: 'Ежедневные задания',
    description: '2 из 4 выполнено',
    image: tasksBoard,
    imageAlt: 'Доска ежедневных заданий',
  },
  {
    title: 'Награды',
    description: '4 бонуса доступны',
    image: awardCup,
    imageAlt: 'Кубок наград',
  },
  {
    title: 'Лидерборд',
    description: 'Ваше место: 18',
    image: leaderboard,
    imageAlt: 'Третье, первое и второе места',
  },
  {
    title: 'Дни активности',
    description: '5 дней на этой неделе',
    image: activityCalendar,
    imageAlt: 'Календарь активности',
  },
];

export const MainPage = () => {
  return (
    <div className={styles.page}>
      <Header petName="Коробыш" />
      <main className={styles.content}>
        <section className={styles.petScene} aria-label="Питомец Коробыш">
          <div className={styles.balance}><i />380 листьев</div>
          <img src={petScene} alt="Коробыш рядом с растением" />
        </section>

        <section className={styles.progressCard} aria-label="Прогресс питомца">
          <div>
            <span>Уровень питомца</span>
            <strong>3/10</strong>
          </div>
          <div className={styles.progressInfo}>
            <b>380 / 560 листьев</b>
            <span>Следующая награда: 100<br />бонусов</span>
          </div>
          <div className={styles.progressBar}><i /></div>
        </section>

        <button className={styles.chestButton} type="button" disabled>
          <strong>Открыть сундук</strong>
          <span>Разблокируется на 10 уровне</span>
        </button>

        <section className={styles.cards} aria-label="Возможности">
          {cards.map((card) => (
            <article className={styles.card} key={card.title}>
              <h2>{card.title}</h2>
              <p>{card.description}</p>
              <img src={card.image} alt={card.imageAlt} />
            </article>
          ))}
        </section>

        <article className={styles.wideCard}>
          <div>
            <h2>Сводка дня</h2>
            <p>3 задания · +140 листьев · место 18</p>
          </div>
          <img src={dailySummary} alt="График сводки дня" />
        </article>

        <article className={styles.wideCard}>
          <div>
            <h2>Правила</h2>
            <p>Как работают листья, уровни и награды</p>
          </div>
          <img src={ruleBook} alt="Книга правил" />
        </article>
      </main>
    </div>
  );
};
