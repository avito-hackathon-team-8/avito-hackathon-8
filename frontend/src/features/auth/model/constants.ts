export const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const OTP_LENGTH = 8;

export type AuthStep = "welcome" | "email" | "code";
