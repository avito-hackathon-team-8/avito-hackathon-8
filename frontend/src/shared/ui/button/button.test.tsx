import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { Button } from './button';

describe('Button', () => {
  it('рендерит кнопку, передаёт атрибуты и обрабатывает нажатие', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();

    render(
      <Button type="submit" variant="primary" onClick={onClick}>
        Сохранить
      </Button>,
    );

    const button = screen.getByRole('button', { name: 'Сохранить' });

    expect(button).toHaveAttribute('type', 'submit');
    await user.click(button);
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('не вызывает обработчик у отключённой кнопки', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();

    render(
      <Button disabled onClick={onClick}>
        Недоступно
      </Button>,
    );

    const button = screen.getByRole('button', { name: 'Недоступно' });

    expect(button).toBeDisabled();
    await user.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });

  it('может быть ссылкой и передаёт её атрибуты', () => {
    render(
      <Button as="a" href="/profile" target="_blank">
        Профиль
      </Button>,
    );

    const link = screen.getByRole('link', { name: 'Профиль' });

    expect(link).toHaveAttribute('href', '/profile');
    expect(link).toHaveAttribute('target', '_blank');
  });
});
