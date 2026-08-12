import { useEffect, useRef } from 'react';

import { usePetProfile } from './use-pet-profile';

interface IUseSceneParams {
  backgroundSrc: string;
  characterSrc: string;
  boxSrc: string;
}

const loadImage = (src: string): Promise<HTMLImageElement> => {
  return new Promise((resolve, reject) => {
    const image = new Image();

    image.src = src;

    image.onload = () => resolve(image);
    image.onerror = reject;
  });
};

const LEVEL_FOR_OPEN_CAT = 2;
const MAX_PET_LEVEL = 10;
const CHARACTER_BASE_WIDTH = 120;
const CHARACTER_INITIAL_SCALE = 0.55;
const CHARACTER_MAX_SCALE = 0.75;
const CHARACTER_VERTICAL_OFFSET = -8;
const CHARACTER_FRAME_COUNT = 3;
const CHARACTER_CONTENT_TOP_RATIO = 0.24;
const CHARACTER_CONTENT_HEIGHT_RATIO = 0.44;
const CHARACTER_BOTTOM_TRANSPARENT_PIXELS = [0, 2, 1] as const;

const getCharacterFrameIndex = (progress: number) => {
  if (progress >= 80) {
    return 2;
  }

  if (progress >= 35) {
    return 1;
  }

  return 0;
};

const getCharacterScale = (level: number) => {
  const clampedLevel = Math.min(Math.max(level, LEVEL_FOR_OPEN_CAT), MAX_PET_LEVEL);
  const levelProgress = (clampedLevel - LEVEL_FOR_OPEN_CAT) / (MAX_PET_LEVEL - LEVEL_FOR_OPEN_CAT);

  return CHARACTER_INITIAL_SCALE + (CHARACTER_MAX_SCALE - CHARACTER_INITIAL_SCALE) * levelProgress;
};

export const useScene = ({ backgroundSrc, characterSrc, boxSrc }: IUseSceneParams) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { data } = usePetProfile();
  const level = data?.level;
  const happiness = data?.happiness;

  useEffect(() => {
    if (!level || happiness === undefined) return;

    const canvas = canvasRef.current;

    if (!canvas) {
      return;
    }

    const ctx = canvas.getContext('2d');

    if (!ctx) {
      return;
    }

    let destroyed = false;

    const render = async () => {
      const [background, character, box] = await Promise.all([
        loadImage(backgroundSrc),
        loadImage(characterSrc),
        loadImage(boxSrc),
      ]);

      if (destroyed) {
        return;
      }

      const rect = canvas.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;

      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;

      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

      ctx.clearRect(0, 0, rect.width, rect.height);

      ctx.drawImage(background, 0, 0, rect.width, rect.height);

      if (level >= LEVEL_FOR_OPEN_CAT) {
        const frameWidth = character.width / CHARACTER_FRAME_COUNT;
        const frameIndex = getCharacterFrameIndex(happiness);
        const sourceY = character.height * CHARACTER_CONTENT_TOP_RATIO;
        const sourceHeight = character.height * CHARACTER_CONTENT_HEIGHT_RATIO;
        const characterWidth = CHARACTER_BASE_WIDTH * getCharacterScale(level);
        const characterHeight = sourceHeight * (characterWidth / frameWidth);
        const bottomTransparentOffset =
          CHARACTER_BOTTOM_TRANSPARENT_PIXELS[frameIndex] * (characterHeight / sourceHeight);

        return ctx.drawImage(
          character,
          frameWidth * frameIndex,
          sourceY,
          frameWidth,
          sourceHeight,
          (rect.width - characterWidth) / 2,
          rect.height - characterHeight + CHARACTER_VERTICAL_OFFSET + bottomTransparentOffset,
          characterWidth,
          characterHeight,
        );
      }

      const boxWidth = 80;
      const boxHeight = box.height * (boxWidth / box.width);

      if (level < LEVEL_FOR_OPEN_CAT) {
        ctx.drawImage(
          box,
          (rect.width - boxWidth) / 2,
          rect.height - boxHeight,
          boxWidth,
          boxHeight,
        );
      }
    };

    render();

    return () => {
      destroyed = true;
    };
  }, [backgroundSrc, boxSrc, characterSrc, happiness, level]);

  return canvasRef;
};
