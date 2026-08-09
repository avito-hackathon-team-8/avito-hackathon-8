import type { SVGProps } from 'react';

import type { TRewardCategory } from '../../api/rewards';

interface TRewardsIconsProps extends SVGProps<SVGSVGElement> {
  variant: TRewardCategory;
}

export const RewardsIcons = ({ variant, ...props }: TRewardsIconsProps) => {
  if (variant === 'AVITO_BONUS') {
    return (
      <svg
        width="40"
        height="40"
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path
          d="M20 40C31.0457 40 40 31.0457 40 20C40 8.9543 31.0457 0 20 0C8.9543 0 0 8.9543 0 20C0 31.0457 8.9543 40 20 40Z"
          fill="#2F9CFF"
        />

        <path
          d="M25 17C28.866 17 32 13.866 32 10C32 6.13401 28.866 3 25 3C21.134 3 18 6.13401 18 10C18 13.866 21.134 17 25 17Z"
          fill="#62C8FF"
        />

        <path
          d="M29.5 29C32.5376 29 35 26.5376 35 23.5C35 20.4624 32.5376 18 29.5 18C26.4624 18 24 20.4624 24 23.5C24 26.5376 26.4624 29 29.5 29Z"
          fill="#FF0000"
        />

        <path
          d="M15 33C18.866 33 22 29.866 22 26C22 22.134 18.866 19 15 19C11.134 19 8 22.134 8 26C8 29.866 11.134 33 15 33Z"
          fill="#74D657"
        />

        <path
          d="M12.5 16C14.433 16 16 14.433 16 12.5C16 10.567 14.433 9 12.5 9C10.567 9 9 10.567 9 12.5C9 14.433 10.567 16 12.5 16Z"
          fill="#7A6FF0"
        />
      </svg>
    );
  }

  if (variant === 'FREE_DELIVERY') {
    return (
      <svg
        width="48"
        height="41"
        viewBox="0 0 48 41"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path d="M0 9H31V33H0V9Z" fill="#64BDF7" />
        <path d="M31 16H40L48 24V33H31V16Z" fill="#3096E6" />

        <path
          d="M12 41C15.3137 41 18 38.3137 18 35C18 31.6863 15.3137 29 12 29C8.68629 29 6 31.6863 6 35C6 38.3137 8.68629 41 12 41Z"
          fill="#385D76"
        />

        <path
          d="M38 41C41.3137 41 44 38.3137 44 35C44 31.6863 41.3137 29 38 29C34.6863 29 32 31.6863 32 35C32 38.3137 34.6863 41 38 41Z"
          fill="#385D76"
        />

        <path
          d="M12 37.5C13.3807 37.5 14.5 36.3807 14.5 35C14.5 33.6193 13.3807 32.5 12 32.5C10.6193 32.5 9.5 33.6193 9.5 35C9.5 36.3807 10.6193 37.5 12 37.5Z"
          fill="#D8F0FF"
        />

        <path
          d="M38 37.5C39.3807 37.5 40.5 36.3807 40.5 35C40.5 33.6193 39.3807 32.5 38 32.5C36.6193 32.5 35.5 33.6193 35.5 35C35.5 36.3807 36.6193 37.5 38 37.5Z"
          fill="#D8F0FF"
        />

        <path d="M38 0L47 4V12C47 19 42.5 23 38 25C33.5 23 29 19 29 12V4L38 0Z" fill="#6BCB55" />

        <path
          d="M34 12L37 15L43 8"
          stroke="white"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  }

  if (variant === 'FREE_PROMOTION') {
    return (
      <svg
        width="46"
        height="51"
        viewBox="0 0 46 51"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path d="M0 21L32 9V39L0 27V21Z" fill="#6BCBFF" />

        <path d="M32 13C38 15 42 19 44 24C42 29 38 33 32 35V13Z" fill="#2A89D8" />

        <path
          d="M6.94116 29.4539L8.60139 44.5414C9.13097 47.5902 11.7886 49.9878 15.8921 49.6494L20.1482 48.3469L16.5126 32.9464L6.94116 29.4539Z"
          fill="#4D6F90"
        />

        <path
          d="M38 16C42.4183 16 46 12.4183 46 8C46 3.58172 42.4183 0 38 0C33.5817 0 30 3.58172 30 8C30 12.4183 33.5817 16 38 16Z"
          fill="#74D657"
        />

        <path d="M35 10L41 6" stroke="white" strokeWidth="2.5" strokeLinecap="round" />

        <path
          d="M35 7.70005C35.9389 7.70005 36.7001 6.93893 36.7001 6.00005C36.7001 5.06116 35.9389 4.30005 35 4.30005C34.0612 4.30005 33.3 5.06116 33.3 6.00005C33.3 6.93893 34.0612 7.70005 35 7.70005Z"
          fill="white"
        />

        <path
          d="M41 11.7001C41.9389 11.7001 42.7001 10.9389 42.7001 10C42.7001 9.06117 41.9389 8.30005 41 8.30005C40.0612 8.30005 39.3 9.06117 39.3 10C39.3 10.9389 40.0612 11.7001 41 11.7001Z"
          fill="white"
        />
      </svg>
    );
  }

  if (variant === 'PROMOTION_DISCOUNT') {
    return (
      <svg
        width="40"
        height="38"
        viewBox="0 0 40 38"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path
          d="M0 6C0 2.7 2.7 0 6 0H27L40 13V32C40 35.3 37.3 38 34 38H6C2.7 38 0 35.3 0 32V6Z"
          fill="#FF5C9A"
        />

        <path
          d="M25 16C27.2091 16 29 14.2091 29 12C29 9.79086 27.2091 8 25 8C22.7909 8 21 9.79086 21 12C21 14.2091 22.7909 16 25 16Z"
          fill="white"
        />

        <path
          d="M13 30C15.2091 30 17 28.2091 17 26C17 23.7909 15.2091 22 13 22C10.7909 22 9 23.7909 9 26C9 28.2091 10.7909 30 13 30Z"
          fill="white"
        />

        <path d="M27 25L11 13" stroke="white" strokeWidth="4" strokeLinecap="round" />

        <path
          d="M33 10C34.6569 10 36 8.65685 36 7C36 5.34315 34.6569 4 33 4C31.3431 4 30 5.34315 30 7C30 8.65685 31.3431 10 33 10Z"
          fill="#FFD15A"
        />
      </svg>
    );
  }

  if (variant === 'DELIVERY_DISCOUNT') {
    return (
      <svg
        width="48"
        height="47"
        viewBox="0 0 48 47"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
      >
        <path
          d="M0.620605 11.5L22.6206 1.5L44.6206 11.5V36.5L22.6206 45.5L0.620605 36.5V11.5Z"
          fill="#F6C759"
        />

        <path
          d="M0.620605 12.5L22.6206 22.5L44.6206 12.5"
          stroke="#E2A93B"
          strokeWidth="3"
          strokeLinejoin="round"
        />

        <path
          d="M1.62061 12.5L22.6206 1.5L43.6206 12.5"
          stroke="#E2A93B"
          strokeWidth="3"
          strokeLinejoin="round"
        />

        <path
          d="M0.620605 36.5L22.6206 45.5L44.6206 36.5"
          stroke="#E2A93B"
          strokeWidth="3"
          strokeLinejoin="round"
        />

        <path d="M22.6206 22.5V45.5" stroke="#E2A93B" strokeWidth="3" />

        <path d="M1.62061 12.5V37.5" stroke="#E2A93B" strokeWidth="3" />

        <path d="M43.6206 11.5V37.5" stroke="#E2A93B" strokeWidth="3" />

        <path
          d="M14.6206 7.5L33.6206 16.5"
          stroke="#FFF1B8"
          strokeWidth="5"
          strokeLinecap="round"
        />

        <path
          d="M31.6206 16.5C31.6206 13.2 34.3206 10.5 37.6206 10.5H40.6206L47.6206 17.5V31.5C47.6206 34.8 44.9206 37.5 41.6206 37.5H34.6206C31.3206 37.5 28.6206 34.8 28.6206 31.5V19.5L31.6206 16.5Z"
          fill="#72D158"
        />

        <path
          d="M40.1206 24.5C41.5013 24.5 42.6206 23.3807 42.6206 22C42.6206 20.6193 41.5013 19.5 40.1206 19.5C38.7399 19.5 37.6206 20.6193 37.6206 22C37.6206 23.3807 38.7399 24.5 40.1206 24.5Z"
          fill="white"
        />

        <path
          d="M35.1206 33.5C36.5013 33.5 37.6206 32.3807 37.6206 31C37.6206 29.6193 36.5013 28.5 35.1206 28.5C33.7399 28.5 32.6206 29.6193 32.6206 31C32.6206 32.3807 33.7399 33.5 35.1206 33.5Z"
          fill="white"
        />

        <path
          d="M42.6206 30.5L34.6206 23.5"
          stroke="white"
          strokeWidth="2.5"
          strokeLinecap="round"
        />
      </svg>
    );
  }

  return (
    <svg
      width="40"
      height="40"
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M20 40C31.0457 40 40 31.0457 40 20C40 8.9543 31.0457 0 20 0C8.9543 0 0 8.9543 0 20C0 31.0457 8.9543 40 20 40Z"
        fill="#2F9CFF"
      />

      <path
        d="M25 17C28.866 17 32 13.866 32 10C32 6.13401 28.866 3 25 3C21.134 3 18 6.13401 18 10C18 13.866 21.134 17 25 17Z"
        fill="#62C8FF"
      />

      <path
        d="M29.5 29C32.5376 29 35 26.5376 35 23.5C35 20.4624 32.5376 18 29.5 18C26.4624 18 24 20.4624 24 23.5C24 26.5376 26.4624 29 29.5 29Z"
        fill="#FF0000"
      />

      <path
        d="M15 33C18.866 33 22 29.866 22 26C22 22.134 18.866 19 15 19C11.134 19 8 22.134 8 26C8 29.866 11.134 33 15 33Z"
        fill="#74D657"
      />

      <path
        d="M12.5 16C14.433 16 16 14.433 16 12.5C16 10.567 14.433 9 12.5 9C10.567 9 9 10.567 9 12.5C9 14.433 10.567 16 12.5 16Z"
        fill="#7A6FF0"
      />
    </svg>
  );
};
