import { Accordion } from '@/shared/ui/accordion';
import { Typography } from '@/shared/ui/typography';

import type { TRule } from '../../model/rules';

import styles from './rules-panel-content.module.scss';

interface IRulesPanelContentProps {
  rules: TRule[];
}

export const RulesPanelContent = ({ rules }: IRulesPanelContentProps) => {
  return (
    <section className={styles.panel}>
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
