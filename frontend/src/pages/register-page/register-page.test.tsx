import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { sessionStorageKeysMap } from '@/shared/lib';
import { renderWithProviders } from '@/test/render-with-providers';

import { RegisterPage } from './register-page';

const mocks = vi.hoisted(() => ({
  requestOtp: vi.fn(),
  verifyOtp: vi.fn(),
  getCurrentUser: vi.fn(),
}));

vi.mock('@/features/auth/api/auth', () => ({
  requestOtp: mocks.requestOtp,
  verifyOtp: mocks.verifyOtp,
}));

vi.mock('@/entities/user', () => ({
  getCurrentUser: mocks.getCurrentUser,
  userQueryKeys: {
    current: () => ['user', 'current'],
  },
}));

describe('RegisterPage', () => {
  beforeEach(() => {
    mocks.requestOtp.mockReset().mockResolvedValue({ sent: true });
    mocks.verifyOtp.mockReset().mockResolvedValue({
      token: 'token-1',
      record: { id: 'user-1', email: 'user@example.com', verified: true },
    });
    mocks.getCurrentUser.mockReset().mockResolvedValue({
      id: 'user-1',
      email: 'user@example.com',
      verified: true,
    });
  });

  it('проходит путь от приветствия до успешной авторизации', async () => {
    const { user } = renderWithProviders(
      <Routes>
        <Route path="/auth" element={<RegisterPage />} />
        <Route path="/" element={<p>Главная страница</p>} />
      </Routes>,
      { route: '/auth' },
    );

    await user.click(screen.getByRole('button', { name: 'Начать' }));
    await user.type(screen.getByRole('textbox', { name: 'Email' }), 'user@example.com');
    await user.click(screen.getByRole('button', { name: 'Получить код' }));

    expect(await screen.findByRole('heading', { name: 'Код из письма' })).toBeInTheDocument();
    expect(mocks.requestOtp).toHaveBeenCalledWith('user@example.com');

    await user.type(screen.getByRole('textbox', { name: 'Код из письма' }), '12345678');
    await user.click(screen.getByRole('button', { name: 'Войти' }));

    expect(await screen.findByText('Главная страница')).toBeInTheDocument();
    expect(mocks.verifyOtp).toHaveBeenCalledWith('user@example.com', '12345678');
    expect(sessionStorage.getItem(sessionStorageKeysMap.authToken)).toBe('token-1');
  });
});
