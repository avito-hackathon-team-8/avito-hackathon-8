import { fireEvent, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createTestQueryClient, renderWithProviders } from '@/test/render-with-providers';

import { PetNameModal } from './pet-name-modal';

const mocks = vi.hoisted(() => ({
  usePetProfile: vi.fn(),
  updatePetName: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  gamificationProfileKeys: {
    pet: () => ['app', 'gamification-profile', 'pet'],
  },
  updatePetName: mocks.updatePetName,
  usePetProfile: mocks.usePetProfile,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const petWithoutName = {
  name: '',
  level: 1,
  leaves: 0,
  nextLevelTargetLeaves: 100,
  chestPrice: 100,
};

describe('PetNameModal', () => {
  beforeEach(() => {
    const portalRoot = document.createElement('div');
    portalRoot.id = 'app-modal-root';
    document.body.append(portalRoot);

    mocks.usePetProfile.mockReset().mockReturnValue({ data: petWithoutName });
    mocks.updatePetName.mockReset();
    mocks.toastError.mockReset();
  });

  afterEach(() => {
    document.getElementById('app-modal-root')?.remove();
    document.body.style.overflow = '';
  });

  it('не открывается без профиля или когда имя уже задано', () => {
    mocks.usePetProfile.mockReturnValue({ data: undefined });
    const { rerender } = renderWithProviders(<PetNameModal isOpen />);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

    mocks.usePetProfile.mockReturnValue({ data: { ...petWithoutName, name: 'Листик' } });
    rerender(<PetNameModal isOpen />);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('сохраняет нормализованное имя и обновляет кэш питомца', async () => {
    const updatedPet = { ...petWithoutName, name: 'Листик' };
    mocks.updatePetName.mockResolvedValue(updatedPet);
    const queryClient = createTestQueryClient();
    const { user } = renderWithProviders(<PetNameModal isOpen />, { queryClient });

    await user.type(screen.getByRole('textbox', { name: 'Имя питомца' }), '  Листик  ');
    await user.click(screen.getByRole('button', { name: 'Сохранить' }));

    expect(mocks.updatePetName.mock.calls[0]?.[0]).toBe('Листик');
    expect(queryClient.getQueryData(['app', 'gamification-profile', 'pet'])).toEqual(updatedPet);
  });

  it('проверяет ограничение длины перед запросом', async () => {
    const { user } = renderWithProviders(<PetNameModal isOpen />);
    const input = screen.getByRole('textbox', { name: 'Имя питомца' });

    fireEvent.change(input, { target: { value: 'а'.repeat(21) } });
    await user.click(screen.getByRole('button', { name: 'Сохранить' }));

    expect(screen.getByText('Имя должно содержать от 1 до 20 символов')).toBeInTheDocument();
    expect(mocks.updatePetName).not.toHaveBeenCalled();
  });

  it('показывает серверную ошибку пользователю', async () => {
    mocks.updatePetName.mockRejectedValue(new Error('Имя уже занято'));
    const { user } = renderWithProviders(<PetNameModal isOpen />);

    await user.type(screen.getByRole('textbox', { name: 'Имя питомца' }), 'Листик');
    await user.click(screen.getByRole('button', { name: 'Сохранить' }));

    expect(await screen.findByText('Имя уже занято')).toBeInTheDocument();
    expect(mocks.toastError).toHaveBeenCalledWith('Не удалось сохранить имя питомца');
  });
});
