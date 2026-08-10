import { screen } from '@testing-library/react';
import { Route, Routes } from 'react-router';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { renderWithProviders } from '@/test/render-with-providers';

import { GuestOnly } from './guest-only';

const useCurrentUserMock = vi.hoisted(() => vi.fn());

vi.mock('@/entities/user', () => ({
  useCurrentUser: useCurrentUserMock,
}));

const TestRoutes = () => (
  <Routes>
    <Route path="/" element={<p>Главная страница</p>} />
    <Route path="/auth" element={<GuestOnly />}>
      <Route index element={<p>Страница входа</p>} />
    </Route>
  </Routes>
);

describe('GuestOnly', () => {
  beforeEach(() => {
    useCurrentUserMock.mockReset();
  });

  it('ничего не показывает во время загрузки пользователя', () => {
    useCurrentUserMock.mockReturnValue({ data: undefined, isLoading: true });

    const { container } = renderWithProviders(<TestRoutes />, { route: '/auth' });
    expect(container).toBeEmptyDOMElement();
  });

  it('показывает страницу входа гостю', () => {
    useCurrentUserMock.mockReturnValue({ data: undefined, isLoading: false });

    renderWithProviders(<TestRoutes />, { route: '/auth' });
    expect(screen.getByText('Страница входа')).toBeInTheDocument();
  });

  it('перенаправляет авторизованного пользователя на главную', async () => {
    useCurrentUserMock.mockReturnValue({
      data: { id: 'user-1', email: 'user@example.com', verified: true },
      isLoading: false,
    });

    renderWithProviders(<TestRoutes />, { route: '/auth' });
    expect(await screen.findByText('Главная страница')).toBeInTheDocument();
    expect(screen.queryByText('Страница входа')).not.toBeInTheDocument();
  });
});
