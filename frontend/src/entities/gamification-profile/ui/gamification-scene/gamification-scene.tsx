import { useScene } from '../../model/use-scene';
import backgroundSrc from '../assets/background.webp';
import boxSrc from '../assets/box.webp';
import characterSrc from '../assets/pet.webp';

import styles from './gamification-scene.module.scss';

export const GamificationScene = () => {
  const canvasRef = useScene({ backgroundSrc, characterSrc, boxSrc });

  return <canvas className={styles.scene} ref={canvasRef} />;
};
