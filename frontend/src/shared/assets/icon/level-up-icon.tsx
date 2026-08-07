import type { SVGProps } from 'react';

export const LevelUpIcon = ({ ...props }: SVGProps<SVGSVGElement>) => {
  return (
    <svg
      width="46"
      height="41"
      viewBox="0 0 46 41"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M9 0H37V10C37 20 31 27 23 30C15 27 9 20 9 10V0Z" fill="#FFC94A" />
      <path d="M9 4H2C2 13 6 18 13 20" stroke="#E7A82C" strokeWidth="4" strokeLinecap="round" />
      <path d="M37 4H44C44 13 40 18 33 20" stroke="#E7A82C" strokeWidth="4" strokeLinecap="round" />
      <path
        d="M24 29H22C20.8954 29 20 29.8954 20 31V34C20 35.1046 20.8954 36 22 36H24C25.1046 36 26 35.1046 26 34V31C26 29.8954 25.1046 29 24 29Z"
        fill="#E7A82C"
      />
      <path
        d="M30 35H16C14.3431 35 13 36.3431 13 38C13 39.6569 14.3431 41 16 41H30C31.6569 41 33 39.6569 33 38C33 36.3431 31.6569 35 30 35Z"
        fill="#F4B63B"
      />
      <path d="M23 19V6" stroke="white" strokeWidth="4" strokeLinecap="round" />
      <path
        d="M17 12L23 6L29 12"
        stroke="white"
        strokeWidth="4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
