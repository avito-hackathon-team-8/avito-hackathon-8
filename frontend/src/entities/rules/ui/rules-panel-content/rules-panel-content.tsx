import { Accordion } from '@/shared/ui/accordion';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import type { TRule } from '../../model/rules';

import styles from './rules-panel-content.module.scss';

interface IRulesPanelContentProps {
  rules: TRule[];
  onStartTutorial?: () => void;
}

export const RulesPanelContent = ({ rules, onStartTutorial }: IRulesPanelContentProps) => {
  return (
    <section className={styles.panel}>
      {onStartTutorial && (
        <Button
          className={styles.panel__tutorialButton}
          isFullWidth
          variant="primary"
          onClick={onStartTutorial}
        >
          Пройти базовое обучение
        </Button>
      )}

      {rules.map(({ id, title, paragraphs }) => (
        <Accordion key={id} title={title} classNameBody={styles.panel__bodyParagraphs}>
          {paragraphs.map((text) => (
            <Typography variant="caption" key={text}>
              {text}
            </Typography>
          ))}
        </Accordion>
      ))}
    </section>
  );
};
