import type { SVGProps } from 'react';

import type { TTaskType } from '../../api/tasks';

interface TRewardsIconsProps extends SVGProps<SVGSVGElement> {
  variant: TTaskType;
}

export const DailyTaskIcons = ({ variant, ...props }: TRewardsIconsProps) => {
  if (variant === 'OPEN_NOTIFICATIONS') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <path
          d="M18 9a6 6 0 1 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9Z"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path d="M10 21h4" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
    );
  }

  if (variant === 'VIEW_LISTINGS') {
    return (
      <svg
        width="48"
        height="37"
        viewBox="0 0 48 37"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path
          d="M3.5 8C3.5 3.6 7.1 0 11.5 0H41.5C43.7 0 45.5 1.8 45.5 4V28C45.5 32.4 41.9 36 37.5 36H7.5C5.3 36 3.5 34.2 3.5 32V8Z"
          fill="#6BCBFF"
        />
        <path
          d="M3.5 8C3.5 3.6 7.1 0 11.5 0H19.5V36H7.5C5.3 36 3.5 34.2 3.5 32V8Z"
          fill="#36A5E8"
        />
        <path d="M14.5 10H37.5" stroke="#DFF6FF" strokeWidth="3" strokeLinecap="round" />
        <path d="M14.5 18H32.5" stroke="#DFF6FF" strokeWidth="3" strokeLinecap="round" />
        <path d="M14.5 26H28.5" stroke="#DFF6FF" strokeWidth="3" strokeLinecap="round" />
        <path
          d="M38.5 37C43.4706 37 47.5 33.1944 47.5 28.5C47.5 23.8056 43.4706 20 38.5 20C33.5294 20 29.5 23.8056 29.5 28.5C29.5 33.1944 33.5294 37 38.5 37Z"
          fill="#62D26F"
        />
        <path
          d="M35.5 28.8462L37.6 31L41.5 27"
          stroke="white"
          strokeWidth="4"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path d="M1.5 10H9.5" stroke="#BDEBFF" strokeWidth="3" strokeLinecap="round" />
        <path d="M1.5 26H9.5" stroke="#BDEBFF" strokeWidth="3" strokeLinecap="round" />
      </svg>
    );
  }

  if (variant === 'ADD_TO_FAVORITES') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <path
          d="M20.8 8.2c0 5-8.8 10.3-8.8 10.3S3.2 13.2 3.2 8.2A4.7 4.7 0 0 1 12 5.9a4.7 4.7 0 0 1 8.8 2.3Z"
          stroke="#a169f7"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }

  if (variant === 'PUBLISH_LISTING') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <rect x="4" y="4" width="16" height="16" rx="3" stroke="currentColor" strokeWidth="1.8" />

        <path d="M12 8v8M8 12h8" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
    );
  }

  if (variant === 'BOOST_LISTING') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <path d="M12 20V5" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />

        <path
          d="m7 10 5-5 5 5"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />

        <path d="M6 20h12" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
    );
  }

  if (variant === 'LEAVE_REVIEW') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <path
          d="m12 3 2.7 5.5 6.1.9-4.4 4.3 1 6.1-5.4-2.9-5.4 2.9 1-6.1-4.4-4.3 6.1-.9L12 3Z"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }

  if (variant === 'COMPLETE_DEAL') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <circle cx="12" cy="12" r="8" stroke="currentColor" strokeWidth="1.8" />

        <path
          d="m8.5 12 2.3 2.3 4.8-5"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }

  if (variant === 'ORDER_WITH_DELIVERY') {
    return (
      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" {...props}>
        <path d="M3 7h11v10H3V7Z" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" />

        <path
          d="M14 10h3.5L21 13.5V17h-7v-7Z"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />

        <circle cx="7" cy="18" r="1.5" stroke="currentColor" strokeWidth="1.8" />

        <circle cx="18" cy="18" r="1.5" stroke="currentColor" strokeWidth="1.8" />
      </svg>
    );
  }

  return null;
};
