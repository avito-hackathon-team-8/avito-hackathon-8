import { describe, expect, it, vi } from 'vitest';

import { apiRequest } from './api-request';

describe('apiRequest', () => {
  it('возвращает разобранный JSON успешного ответа', async () => {
    const response = new Response(JSON.stringify({ id: 1, name: 'Листик' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });

    await expect(apiRequest(Promise.resolve(response))).resolves.toEqual({
      id: 1,
      name: 'Листик',
    });
  });

  it('возвращает undefined для ответа 204', async () => {
    const response = new Response(null, { status: 204 });

    await expect(apiRequest(Promise.resolve(response))).resolves.toBeUndefined();
  });

  it('вызывает обработчик и использует текст ошибки из ответа', async () => {
    const onError = vi.fn();
    const response = new Response('Неверный код', { status: 400 });

    await expect(
      apiRequest(Promise.resolve(response), 'Ошибка авторизации', onError),
    ).rejects.toThrow('Неверный код');
    expect(onError).toHaveBeenCalledOnce();
  });

  it('использует резервный текст для пустого ответа с ошибкой', async () => {
    const response = new Response(null, { status: 500 });

    await expect(apiRequest(Promise.resolve(response), 'Сервис недоступен')).rejects.toThrow(
      'Сервис недоступен',
    );
  });
});
