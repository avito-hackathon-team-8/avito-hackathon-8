import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import { Accordion } from './accordion';

describe('Accordion', () => {
  it('открывает и закрывает содержимое по нажатию', async () => {
    const user = userEvent.setup();

    render(<Accordion title="Как получить награду?">Выполните задание</Accordion>);

    const trigger = screen.getByRole('button', { name: 'Как получить награду?' });
    const contentId = trigger.getAttribute('aria-controls');
    const content = document.getElementById(contentId ?? '');

    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(content).toHaveAttribute('aria-hidden', 'true');

    await user.click(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'true');
    expect(content).toHaveAttribute('aria-hidden', 'false');

    await user.click(trigger);
    expect(trigger).toHaveAttribute('aria-expanded', 'false');
    expect(content).toHaveAttribute('aria-hidden', 'true');
  });

  it('учитывает начальное открытое состояние', () => {
    render(
      <Accordion title="Уровни" defaultOpen>
        Список уровней
      </Accordion>,
    );

    expect(screen.getByRole('button', { name: 'Уровни' })).toHaveAttribute('aria-expanded', 'true');
  });
});
