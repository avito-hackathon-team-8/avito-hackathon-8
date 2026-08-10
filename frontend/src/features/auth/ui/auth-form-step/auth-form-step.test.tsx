import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { AuthFormStep } from './auth-form-step';

describe('AuthFormStep', () => {
  it('передаёт введённое значение и отправляет email-форму', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const onSubmit = vi.fn();

    render(
      <AuthFormStep
        variant="email"
        value=""
        error=""
        isSubmitting={false}
        isValid
        onChange={onChange}
        onSubmit={onSubmit}
      />,
    );

    const input = screen.getByRole('textbox', { name: 'Email' });
    const submitButton = screen.getByRole('button', { name: 'Получить код' });

    expect(input).toHaveAttribute('type', 'email');
    await user.type(input, 'a');
    expect(onChange).toHaveBeenCalledWith('a');

    await user.click(submitButton);
    expect(onSubmit).toHaveBeenCalledOnce();
  });

  it('связывает ошибку с полем и запрещает невалидную отправку', () => {
    const onSubmit = vi.fn();

    render(
      <AuthFormStep
        variant="email"
        value="invalid-email"
        error="Введите корректный email"
        isSubmitting={false}
        isValid={false}
        onChange={vi.fn()}
        onSubmit={onSubmit}
      />,
    );

    const input = screen.getByRole('textbox', { name: 'Email' });
    const error = screen.getByText('Введите корректный email');
    const submitButton = screen.getByRole('button', { name: 'Получить код' });

    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(input).toHaveAttribute('aria-describedby', error.id);
    expect(submitButton).toBeDisabled();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('настраивает поле кода и позволяет вернуться к email', async () => {
    const user = userEvent.setup();
    const onBack = vi.fn();

    render(
      <AuthFormStep
        variant="code"
        value="1234"
        error=""
        isSubmitting={false}
        isValid={false}
        onChange={vi.fn()}
        onSubmit={vi.fn()}
        onBack={onBack}
      />,
    );

    const input = screen.getByRole('textbox', { name: 'Код из письма' });

    expect(input).toHaveAttribute('inputmode', 'numeric');
    expect(input).toHaveAttribute('maxlength', '8');

    await user.click(screen.getByRole('button', { name: 'Вернуться к вводу почты' }));
    expect(onBack).toHaveBeenCalledOnce();
  });
});
