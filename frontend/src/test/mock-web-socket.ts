import { vi } from 'vitest';

export class MockWebSocket {
  static instances: MockWebSocket[] = [];

  readonly url: string;

  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;

  close = vi.fn();

  constructor(url: string | URL) {
    this.url = String(url);
    MockWebSocket.instances.push(this);
  }

  emitOpen() {
    this.onopen?.(new Event('open'));
  }

  emitMessage(data: unknown) {
    const message = typeof data === 'string' ? data : JSON.stringify(data);
    this.onmessage?.(new MessageEvent('message', { data: message }));
  }

  emitError() {
    this.onerror?.(new Event('error'));
  }

  emitClose(code = 1006, reason = 'Connection lost') {
    this.onclose?.(
      new CloseEvent('close', {
        code,
        reason,
        wasClean: code === 1000,
      }),
    );
  }

  static reset() {
    MockWebSocket.instances = [];
  }
}
