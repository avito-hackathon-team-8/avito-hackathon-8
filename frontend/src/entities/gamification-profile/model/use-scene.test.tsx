import { act, render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useScene } from './use-scene';

const mocks = vi.hoisted(() => ({
  usePetProfile: vi.fn(),
}));

vi.mock('./use-pet-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

class MockImage {
  static instances: MockImage[] = [];

  onload: (() => void) | null = null;
  onerror: (() => void) | null = null;
  src = '';
  width = 200;
  height = 100;

  constructor() {
    MockImage.instances.push(this);
  }

  emitLoad() {
    this.onload?.();
  }

  static reset() {
    MockImage.instances = [];
  }
}

const context = {
  setTransform: vi.fn(),
  clearRect: vi.fn(),
  drawImage: vi.fn(),
};

const SceneHarness = () => {
  const canvasRef = useScene({
    backgroundSrc: '/background.webp',
    characterSrc: '/pet.webp',
    boxSrc: '/box.webp',
  });

  return <canvas ref={canvasRef} />;
};

const loadSceneImages = async () => {
  await act(async () => {
    MockImage.instances.forEach((image) => image.emitLoad());
    await Promise.resolve();
  });
};

describe('useScene', () => {
  beforeEach(() => {
    MockImage.reset();
    context.setTransform.mockReset();
    context.clearRect.mockReset();
    context.drawImage.mockReset();
    vi.stubGlobal('Image', MockImage);
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(
      () => context as unknown as CanvasRenderingContext2D,
    );
    vi.spyOn(HTMLCanvasElement.prototype, 'getBoundingClientRect').mockReturnValue({
      bottom: 200,
      height: 200,
      left: 0,
      right: 300,
      top: 0,
      width: 300,
      x: 0,
      y: 0,
      toJSON: () => ({}),
    });
    Object.defineProperty(window, 'devicePixelRatio', {
      configurable: true,
      value: 1,
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('рисует коробку на первом уровне', async () => {
    mocks.usePetProfile.mockReturnValue({ data: { level: 1 } });
    render(<SceneHarness />);

    expect(MockImage.instances.map(({ src }) => src)).toEqual([
      '/background.webp',
      '/pet.webp',
      '/box.webp',
    ]);
    await loadSceneImages();

    expect(context.drawImage).toHaveBeenCalledTimes(2);
    expect(context.drawImage).toHaveBeenNthCalledWith(2, MockImage.instances[2], 110, 160, 80, 40);
  });

  it('рисует питомца со второго уровня с начальным масштабом', async () => {
    mocks.usePetProfile.mockReturnValue({ data: { level: 2 } });
    render(<SceneHarness />);
    await loadSceneImages();

    expect(context.drawImage).toHaveBeenNthCalledWith(2, MockImage.instances[1], 114, 172, 72, 36);
  });

  it('ограничивает рост питомца максимальным масштабом десятого уровня', async () => {
    mocks.usePetProfile.mockReturnValue({ data: { level: 15 } });
    render(<SceneHarness />);
    await loadSceneImages();

    expect(context.drawImage).toHaveBeenNthCalledWith(2, MockImage.instances[1], 102, 160, 96, 48);
  });

  it('масштабирует canvas под devicePixelRatio', async () => {
    Object.defineProperty(window, 'devicePixelRatio', {
      configurable: true,
      value: 2,
    });
    mocks.usePetProfile.mockReturnValue({ data: { level: 2 } });
    const { container } = render(<SceneHarness />);
    await loadSceneImages();
    const canvas = container.querySelector('canvas');

    expect(canvas).toHaveProperty('width', 600);
    expect(canvas).toHaveProperty('height', 400);
    expect(context.setTransform).toHaveBeenCalledWith(2, 0, 0, 2, 0, 0);
  });

  it('не рисует изображения, загрузившиеся после unmount', async () => {
    mocks.usePetProfile.mockReturnValue({ data: { level: 2 } });
    const { unmount } = render(<SceneHarness />);

    unmount();
    await loadSceneImages();
    expect(context.drawImage).not.toHaveBeenCalled();
  });
});
