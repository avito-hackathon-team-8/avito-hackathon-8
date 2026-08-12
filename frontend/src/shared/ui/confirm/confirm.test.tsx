import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { Confirm } from './confirm';

describe('Confirm', () => {
  it('показывает предупреждение и вызывает действия', async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    const onCancel = vi.fn();

    render(<Confirm title="Подтвердите действие" onConfirm={onConfirm} onCancel={onCancel} />);

    expect(screen.getByRole('alert')).toHaveTextContent('Подтвердите действие');

    await user.click(screen.getByRole('button', { name: 'Подтвердить' }));
    await user.click(screen.getByRole('button', { name: 'Отменить' }));

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('блокирует действия', () => {
    render(
      <Confirm title="Подтвердите действие" onConfirm={vi.fn()} onCancel={vi.fn()} disabled />,
    );

    expect(screen.getByRole('button', { name: 'Подтвердить' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Отменить' })).toBeDisabled();
  });
});
