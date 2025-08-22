/**
 * Class Name Utility
 * Combines class-variance-authority with clsx for conditional class names
 */

import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * Combines class names with proper Tailwind CSS merging
 * 
 * @param inputs - Class values to combine
 * @returns Merged class string
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}

export default cn