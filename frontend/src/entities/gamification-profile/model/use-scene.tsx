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

export const useScene = ({ backgroundSrc, characterSrc, boxSrc }: IUseSceneParams) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const { data } = usePetProfile();
  const level = data?.level;

  useEffect(() => {
    if (!level) return;

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

      const characterWidth = 120;
      const characterHeight = character.height * (characterWidth / character.width);

      if (level >= LEVEL_FOR_OPEN_CAT) {
        return ctx.drawImage(
          character,
          (rect.width - characterWidth) / 2,
          rect.height - characterHeight,
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
  }, [backgroundSrc, boxSrc, characterSrc, level]);

  return canvasRef;
};
