import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { describeError } from '@/lib/errors';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * A message for an unknown caught value that a person can read and act on.
 */
export function errorMessage(e: unknown): string {
  return describeError(e).message;
}
