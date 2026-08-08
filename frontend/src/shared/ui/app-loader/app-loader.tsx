import styles from './app-loader.module.scss';

export const AppLoader = () => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.loader} role="status" aria-label="Загрузка" />
    </div>
  );
};
