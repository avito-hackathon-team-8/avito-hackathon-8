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
  private imageSrc = '';
  width = 200;
  height = 100;

  constructor() {
    MockImage.instances.push(this);
  }

  get src() {
    return this.imageSrc;
  }

  set src(value: string) {
    this.imageSrc = value;

    if (value === '/pet.webp') {
      this.width = 300;
      this.height = 300;
    }
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

const expectCharacterFrame = (frameIndex: number, width: number, height: number) => {
  const [, sourceX, sourceY, sourceWidth, sourceHeight, x, y, drawnWidth, drawnHeight] =
    context.drawImage.mock.calls[1];

  expect(sourceX).toBe(frameIndex * 100);
  expect(sourceY).toBe(72);
  expect(sourceWidth).toBe(100);
  expect(sourceHeight).toBe(132);
  expect(x).toBeCloseTo((300 - width) / 2);
  const bottomTransparentPixels = [0, 2, 1][frameIndex];
  const bottomTransparentOffset = bottomTransparentPixels * (height / sourceHeight);

  expect(y).toBeCloseTo(200 - height - 8 + bottomTransparentOffset);
  expect(drawnWidth).toBeCloseTo(width);
  expect(drawnHeight).toBeCloseTo(height);
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
    mocks.usePetProfile.mockReturnValue({
      data: { happiness: 50, level: 1 },
    });
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

  it('рисует второй кадр питомца для временного прогресса 50', async () => {
    mocks.usePetProfile.mockReturnValue({
      data: { happiness: 50, level: 2 },
    });
    render(<SceneHarness />);
    await loadSceneImages();

    expectCharacterFrame(1, 66, 87.12);
  });

  it.each([
    [34, 0],
    [35, 1],
    [79, 1],
    [80, 2],
  ])('выбирает кадр %s для счастья %s', async (happiness, frameIndex) => {
    mocks.usePetProfile.mockReturnValue({
      data: { happiness, level: 2 },
    });
    render(<SceneHarness />);
    await loadSceneImages();

    expectCharacterFrame(frameIndex, 66, 87.12);
  });

  it('ограничивает рост питомца масштабом 0.75 на десятом уровне', async () => {
    mocks.usePetProfile.mockReturnValue({
      data: { happiness: 100, level: 15 },
    });
    render(<SceneHarness />);
    await loadSceneImages();

    expectCharacterFrame(2, 90, 118.8);
  });

  it('масштабирует canvas под devicePixelRatio', async () => {
    Object.defineProperty(window, 'devicePixelRatio', {
      configurable: true,
      value: 2,
    });
    mocks.usePetProfile.mockReturnValue({
      data: { happiness: 50, level: 2 },
    });
    const { container } = render(<SceneHarness />);
    await loadSceneImages();
    const canvas = container.querySelector('canvas');

    expect(canvas).toHaveProperty('width', 600);
    expect(canvas).toHaveProperty('height', 400);
    expect(context.setTransform).toHaveBeenCalledWith(2, 0, 0, 2, 0, 0);
  });

  it('не рисует изображения, загрузившиеся после unmount', async () => {
    mocks.usePetProfile.mockReturnValue({
      data: { happiness: 50, level: 2 },
    });
    const { unmount } = render(<SceneHarness />);

    unmount();
    await loadSceneImages();
    expect(context.drawImage).not.toHaveBeenCalled();
  });
});
