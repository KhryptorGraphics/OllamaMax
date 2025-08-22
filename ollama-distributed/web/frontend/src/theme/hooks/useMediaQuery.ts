import { useState, useEffect } from 'react';

/**
 * Hook to check if a media query matches
 */
export const useMediaQuery = (query: string): boolean => {
  const [matches, setMatches] = useState(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia(query).matches;
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const mediaQuery = window.matchMedia(query);
    const handleChange = (e: MediaQueryListEvent) => {
      setMatches(e.matches);
    };

    mediaQuery.addEventListener('change', handleChange);
    
    // Set initial value
    setMatches(mediaQuery.matches);

    return () => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  }, [query]);

  return matches;
};

/**
 * Hook to get current breakpoint
 */
export const useBreakpoint = () => {
  const isXs = useMediaQuery('(min-width: 475px)');
  const isSm = useMediaQuery('(min-width: 640px)');
  const isMd = useMediaQuery('(min-width: 768px)');
  const isLg = useMediaQuery('(min-width: 1024px)');
  const isXl = useMediaQuery('(min-width: 1280px)');
  const is2xl = useMediaQuery('(min-width: 1536px)');

  const current = is2xl ? '2xl' : isXl ? 'xl' : isLg ? 'lg' : isMd ? 'md' : isSm ? 'sm' : isXs ? 'xs' : 'base';

  return {
    current,
    isXs,
    isSm,
    isMd,
    isLg,
    isXl,
    is2xl,
    isMobile: !isSm,
    isTablet: isSm && !isLg,
    isDesktop: isLg
  };
};

/**
 * Hook to check for reduced motion preference
 */
export const usePrefersReducedMotion = (): boolean => {
  return useMediaQuery('(prefers-reduced-motion: reduce)');
};

/**
 * Hook to check for high contrast preference
 */
export const usePrefersHighContrast = (): boolean => {
  return useMediaQuery('(prefers-contrast: high)');
};

/**
 * Hook to detect touch devices
 */
export const useIsTouchDevice = (): boolean => {
  return useMediaQuery('(hover: none) and (pointer: coarse)');
};