import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { WelcomeStep } from './welcome-step';

describe('WelcomeStep', () => {
  it('показывает приглашение и продолжает авторизацию по нажатию', async () => {
    const user = userEvent.setup();
    const onContinue = vi.fn();

    render(<WelcomeStep onContinue={onContinue} />);

    expect(screen.getByRole('heading', { name: 'Привет, это игра «Коробыш»' })).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Начать' }));
    expect(onContinue).toHaveBeenCalledOnce();
  });
});
