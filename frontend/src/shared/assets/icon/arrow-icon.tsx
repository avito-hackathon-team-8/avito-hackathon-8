import type { SVGAttributes } from 'react';

export const ArrowIcon = ({ ...props }: SVGAttributes<SVGSVGElement>) => {
  return (
    <svg
      width="19"
      height="16"
      viewBox="0 0 19 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <g clipPath="url(#clip0_89_3)">
        <path
          d="M7.95455 15.9091L0 7.95454L7.95455 0L9.32174 1.34943L3.69318 6.97798H18.2706V8.9311H3.69318L9.32174 14.5419L7.95455 15.9091Z"
          fill="#0F0F0F"
        />
      </g>
      <defs>
        <clipPath id="clip0_89_3">
          <rect width="19" height="16" fill="white" />
        </clipPath>
      </defs>
    </svg>
  );
};
