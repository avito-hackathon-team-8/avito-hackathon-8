import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from './bottom-panel';
import { BottomPanel } from './bottom-panel';

const renderPanel = (props: Partial<BottomPanelProps> = {}) => {
  return render(
    <BottomPanel
      title="Дни активности"
      description="Награда зависит от календарного дня"
      renderTrigger={(open) => (
        <button type="button" onClick={open}>
          Открыть панель
        </button>
      )}
      {...props}
    >
      <button type="button">Действие в панели</button>
    </BottomPanel>,
  );
};

describe('BottomPanel', () => {
  let shell: HTMLDivElement;
  let appContent: HTMLDivElement;

  beforeEach(() => {
    shell = document.createElement('div');
    appContent = document.createElement('div');
    appContent.dataset.appContent = '';

    const portalRoot = document.createElement('div');
    portalRoot.id = 'app-overlay-root';

    shell.append(appContent, portalRoot);
    document.body.append(shell);

    vi.spyOn(HTMLElement.prototype, 'getClientRects').mockReturnValue([
      {
        bottom: 20,
        height: 20,
        left: 0,
        right: 20,
        top: 0,
        width: 20,
        x: 0,
        y: 0,
        toJSON: () => ({}),
      },
    ] as unknown as DOMRectList);
  });

  afterEach(() => {
    shell.remove();
    document.documentElement.style.overflow = '';
    document.body.style.overflow = '';
  });

  it('открывается, блокирует страницу и переводит фокус внутрь', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderPanel({ onClick });

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));

    const dialog = screen.getByRole('dialog', { name: 'Дни активности' });
    expect(dialog).toHaveTextContent('Награда зависит от календарного дня');
    expect(dialog.parentElement).toHaveAttribute('data-open', 'true');
    expect(onClick).toHaveBeenCalledOnce();
    expect(document.documentElement).toHaveStyle({ overflow: 'hidden' });
    expect(document.body).toHaveStyle({ overflow: 'hidden' });
    expect(appContent.inert).toBe(true);
    expect(screen.getByText('×').closest('button')).toHaveFocus();
  });

  it('вызывает onClick, но не открывается в disabled-состоянии', async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    renderPanel({ disabled: true, onClick });

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));

    expect(onClick).toHaveBeenCalledOnce();
    expect(screen.getByRole('dialog').parentElement).toHaveAttribute('data-open', 'false');
    expect(document.body.style.overflow).toBe('');
  });

  it('закрывается по Escape, снимает блокировки и возвращает фокус', async () => {
    const user = userEvent.setup();
    renderPanel();
    const trigger = screen.getByRole('button', { name: 'Открыть панель' });

    await user.click(trigger);
    await user.keyboard('{Escape}');

    expect(screen.getByRole('dialog').parentElement).toHaveAttribute('data-open', 'false');
    expect(document.documentElement.style.overflow).toBe('');
    expect(document.body.style.overflow).toBe('');
    expect(appContent.inert).toBe(false);
    expect(trigger).toHaveFocus();
  });

  it('замыкает Tab с последнего элемента на первый', async () => {
    const user = userEvent.setup();
    renderPanel();

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));
    const closeButton = screen.getByText('×').closest('button');
    const actionButton = screen.getByRole('button', { name: 'Действие в панели' });

    expect(closeButton).not.toBeNull();
    closeButton!.focus();
    await user.tab();
    expect(actionButton).toHaveFocus();

    await user.tab();
    expect(closeButton).toHaveFocus();
  });

  it('замыкает Shift+Tab с первого элемента на последний', async () => {
    const user = userEvent.setup();
    renderPanel();

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));
    const closeButton = screen.getByText('×').closest('button');
    const actionButton = screen.getByRole('button', { name: 'Действие в панели' });

    expect(closeButton).not.toBeNull();
    closeButton!.focus();
    await user.tab({ shift: true });
    expect(actionButton).toHaveFocus();
  });

  it('возвращает программный фокус снаружи обратно в панель', async () => {
    const user = userEvent.setup();
    const outsideButton = document.createElement('button');
    appContent.append(outsideButton);
    renderPanel();

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));
    outsideButton.focus();

    expect(screen.getByText('×').closest('button')).toHaveFocus();
  });

  it('закрывается по клику на backdrop', async () => {
    const user = userEvent.setup();
    renderPanel();

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));
    const overlay = screen.getByRole('dialog').parentElement;
    expect(overlay).not.toBeNull();

    await user.click(overlay!);
    expect(overlay).toHaveAttribute('data-open', 'false');
  });

  it('не закрывается по backdrop при closeOnBackdrop=false', async () => {
    const user = userEvent.setup();
    renderPanel({ closeOnBackdrop: false });

    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));
    const overlay = screen.getByRole('dialog').parentElement;
    expect(overlay).not.toBeNull();

    await user.click(overlay!);
    expect(overlay).toHaveAttribute('data-open', 'true');
  });

  it('сообщает об отсутствующем portal root', async () => {
    const user = userEvent.setup();
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    document.getElementById('app-overlay-root')?.remove();

    renderPanel();
    await user.click(screen.getByRole('button', { name: 'Открыть панель' }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(consoleError).toHaveBeenCalledWith('BottomPanel: не найден элемент #app-overlay-root');
  });

  it('удаляет keyboard и focus listeners при unmount', () => {
    const addEventListener = vi.spyOn(document, 'addEventListener');
    const removeEventListener = vi.spyOn(document, 'removeEventListener');
    const { unmount } = renderPanel({ isStartOpen: true });

    expect(addEventListener).toHaveBeenCalledWith('keydown', expect.any(Function));
    expect(addEventListener).toHaveBeenCalledWith('focusin', expect.any(Function));

    unmount();
    expect(removeEventListener).toHaveBeenCalledWith('keydown', expect.any(Function));
    expect(removeEventListener).toHaveBeenCalledWith('focusin', expect.any(Function));
  });
});
