import type { SVGAttributes } from 'react';

export const NotificationIcon = ({ ...props }: SVGAttributes<SVGSVGElement>) => {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M6.5 9.75C6.5 6.574 8.962 4 12 4C15.038 4 17.5 6.574 17.5 9.75V13.05C17.5 14.108 17.86 15.133 18.52 15.96L19.25 16.875C19.746 17.497 19.303 18.417 18.507 18.417H5.493C4.697 18.417 4.254 17.497 4.75 16.875L5.48 15.96C6.14 15.133 6.5 14.108 6.5 13.05V9.75Z"
        stroke="#0F0F0F"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M9.75 20C10.27 20.63 11.057 21 12 21C12.943 21 13.73 20.63 14.25 20"
        stroke="#0F0F0F"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
    </svg>
  );
};
