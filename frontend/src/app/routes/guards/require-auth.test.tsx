import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { renderWithProviders } from '@/test/render-with-providers';

import { RequireAuth } from './require-auth';

const useCurrentUserMock = vi.hoisted(() => vi.fn());

vi.mock('@/entities/user', () => ({
  useCurrentUser: useCurrentUserMock,
}));

const TestRoutes = () => (
  <Routes>
    <Route path="/auth" element={<p>Страница входа</p>} />
    <Route path="/private" element={<RequireAuth />}>
      <Route index element={<p>Личный кабинет</p>} />
    </Route>
  </Routes>
);

describe('RequireAuth', () => {
  beforeEach(() => {
    useCurrentUserMock.mockReset();
  });

  it('показывает состояние загрузки', () => {
    useCurrentUserMock.mockReturnValue({ data: undefined, isLoading: true });

    renderWithProviders(<TestRoutes />, { route: '/private' });

    expect(screen.getByText('Загрузка')).toBeInTheDocument();
  });

  it('перенаправляет гостя на страницу входа', async () => {
    useCurrentUserMock.mockReturnValue({ data: undefined, isLoading: false });

    renderWithProviders(<TestRoutes />, { route: '/private' });

    expect(await screen.findByText('Страница входа')).toBeInTheDocument();
    expect(screen.queryByText('Личный кабинет')).not.toBeInTheDocument();
  });

  it('показывает защищённую страницу авторизованному пользователю', () => {
    useCurrentUserMock.mockReturnValue({ data: { id: 'user-1' }, isLoading: false });

    renderWithProviders(<TestRoutes />, { route: '/private' });

    expect(screen.getByText('Личный кабинет')).toBeInTheDocument();
  });
});
