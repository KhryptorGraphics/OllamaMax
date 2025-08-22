export const animation = {
  // Duration tokens
  duration: {
    instant: '0ms',
    fast: '150ms',
    normal: '250ms',
    slow: '350ms',
    slower: '500ms',
    slowest: '750ms'
  },

  // Easing curves
  easing: {
    linear: 'linear',
    ease: 'ease',
    'ease-in': 'cubic-bezier(0.4, 0, 1, 1)',
    'ease-out': 'cubic-bezier(0, 0, 0.2, 1)',
    'ease-in-out': 'cubic-bezier(0.4, 0, 0.2, 1)',
    
    // Custom easing curves
    'bounce-in': 'cubic-bezier(0.68, -0.6, 0.32, 1.6)',
    'bounce-out': 'cubic-bezier(0.68, -0.6, 0.32, 1.6)',
    'back-in': 'cubic-bezier(0.6, -0.28, 0.735, 0.045)',
    'back-out': 'cubic-bezier(0.175, 0.885, 0.32, 1.275)',
    
    // Sharp curves for UI
    'sharp-in': 'cubic-bezier(0.4, 0, 0.6, 1)',
    'sharp-out': 'cubic-bezier(0.4, 0, 0.2, 1)',
    
    // Smooth curves
    'smooth': 'cubic-bezier(0.25, 0.46, 0.45, 0.94)',
    'smooth-in': 'cubic-bezier(0.25, 0.46, 0.45, 0.94)',
    'smooth-out': 'cubic-bezier(0.25, 0.46, 0.45, 0.94)'
  },

  // Scale transforms
  scale: {
    enter: '1.05',
    exit: '0.95',
    press: '0.98',
    hover: '1.02'
  },

  // Translate distances
  translate: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px'
  }
} as const;

// Motion preferences and CSS variables
export const motionTokens = {
  // Respect user's motion preferences
  scale: 'var(--motion-scale, 1.05)',
  duration: 'var(--motion-duration, 250ms)',
  easing: 'var(--motion-easing, cubic-bezier(0.4, 0, 0.2, 1))',
  
  // Reduced motion fallbacks
  'scale-reduced': 'var(--motion-scale, 1.01)',
  'duration-reduced': 'var(--motion-duration, 150ms)',
  'easing-reduced': 'var(--motion-easing, ease-out)'
} as const;

// Common animation presets
export const animations = {
  // Fade animations
  fadeIn: {
    from: { opacity: 0 },
    to: { opacity: 1 }
  },
  fadeOut: {
    from: { opacity: 1 },
    to: { opacity: 0 }
  },

  // Slide animations
  slideInUp: {
    from: { 
      opacity: 0, 
      transform: `translateY(${animation.translate.md})` 
    },
    to: { 
      opacity: 1, 
      transform: 'translateY(0)' 
    }
  },
  slideInDown: {
    from: { 
      opacity: 0, 
      transform: `translateY(-${animation.translate.md})` 
    },
    to: { 
      opacity: 1, 
      transform: 'translateY(0)' 
    }
  },
  slideInLeft: {
    from: { 
      opacity: 0, 
      transform: `translateX(-${animation.translate.md})` 
    },
    to: { 
      opacity: 1, 
      transform: 'translateX(0)' 
    }
  },
  slideInRight: {
    from: { 
      opacity: 0, 
      transform: `translateX(${animation.translate.md})` 
    },
    to: { 
      opacity: 1, 
      transform: 'translateX(0)' 
    }
  },

  // Scale animations
  scaleIn: {
    from: { 
      opacity: 0, 
      transform: `scale(${animation.scale.exit})` 
    },
    to: { 
      opacity: 1, 
      transform: 'scale(1)' 
    }
  },
  scaleOut: {
    from: { 
      opacity: 1, 
      transform: 'scale(1)' 
    },
    to: { 
      opacity: 0, 
      transform: `scale(${animation.scale.exit})` 
    }
  },

  // Bounce animations
  bounceIn: {
    from: { 
      opacity: 0, 
      transform: 'scale(0.3)' 
    },
    '50%': { 
      transform: 'scale(1.05)' 
    },
    '70%': { 
      transform: 'scale(0.9)' 
    },
    to: { 
      opacity: 1, 
      transform: 'scale(1)' 
    }
  },

  // Loading animations
  spin: {
    from: { transform: 'rotate(0deg)' },
    to: { transform: 'rotate(360deg)' }
  },
  pulse: {
    '0%, 100%': { opacity: 1 },
    '50%': { opacity: 0.5 }
  },
  ping: {
    '75%, 100%': {
      transform: 'scale(2)',
      opacity: 0
    }
  }
} as const;

// Keyframe definitions for CSS-in-JS
export const keyframes = Object.entries(animations).reduce((acc, [name, frames]) => {
  acc[name] = frames;
  return acc;
}, {} as Record<string, any>);

export type Animation = typeof animation;
export type MotionTokens = typeof motionTokens;
export type Animations = typeof animations;