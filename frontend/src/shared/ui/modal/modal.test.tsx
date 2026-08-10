import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { Modal } from './modal';

describe('Modal', () => {
  beforeEach(() => {
    const portalRoot = document.createElement('div');
    portalRoot.id = 'app-modal-root';
    document.body.append(portalRoot);
  });

  afterEach(() => {
    document.body.style.overflow = '';
    document.getElementById('app-modal-root')?.remove();
  });

  it('показывает содержимое, блокирует прокрутку и закрывается по Escape', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const { rerender } = render(
      <Modal isOpen onClose={onClose}>
        Содержимое модального окна
      </Modal>,
    );

    expect(screen.getByRole('dialog')).toHaveTextContent('Содержимое модального окна');
    expect(document.body).toHaveStyle({ overflow: 'hidden' });

    await user.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalledOnce();

    rerender(
      <Modal isOpen={false} onClose={onClose}>
        Содержимое модального окна
      </Modal>,
    );
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(document.body.style.overflow).toBe('');
  });

  it('закрывается по оверлею, но не по содержимому', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();

    render(
      <Modal isOpen onClose={onClose}>
        Содержимое
      </Modal>,
    );

    const dialog = screen.getByRole('dialog');
    const overlay = dialog.parentElement;

    await user.click(dialog);
    expect(onClose).not.toHaveBeenCalled();

    expect(overlay).not.toBeNull();
    await user.click(overlay!);
    expect(onClose).toHaveBeenCalledOnce();
  });
});
