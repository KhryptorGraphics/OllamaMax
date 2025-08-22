/**
 * Design Tokens: Transitions
 * Comprehensive animation and transition system with performance optimization and accessibility
 */

// Base timing functions (easing curves)
export const easingCurves = {
  // Standard easing
  linear: 'cubic-bezier(0, 0, 1, 1)',
  ease: 'cubic-bezier(0.25, 0.1, 0.25, 1)',
  easeIn: 'cubic-bezier(0.42, 0, 1, 1)',
  easeOut: 'cubic-bezier(0, 0, 0.58, 1)',
  easeInOut: 'cubic-bezier(0.42, 0, 0.58, 1)',
  
  // Material Design easing
  standard: 'cubic-bezier(0.4, 0, 0.2, 1)',
  decelerate: 'cubic-bezier(0, 0, 0.2, 1)',
  accelerate: 'cubic-bezier(0.4, 0, 1, 1)',
  sharp: 'cubic-bezier(0.4, 0, 0.6, 1)',
  
  // Expressive easing
  spring: 'cubic-bezier(0.175, 0.885, 0.32, 1.275)',
  bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
  elastic: 'cubic-bezier(0.68, -0.6, 0.32, 1.6)',
  
  // Natural motion
  entrance: 'cubic-bezier(0, 0, 0.2, 1)',
  exit: 'cubic-bezier(0.4, 0, 1, 1)',
  emphasized: 'cubic-bezier(0.2, 0, 0, 1)'
} as const

// Base duration values
export const durations = {
  instant: '0ms',
  fast: '150ms',
  normal: '200ms',
  medium: '300ms',
  slow: '500ms',
  slower: '700ms',
  slowest: '1000ms'
} as const

// Semantic transition tokens
export const semanticTransitions = {
  // Interactive element transitions
  interactive: {
    // Button transitions
    button: {
      background: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'background-color'
      },
      transform: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'transform'
      },
      shadow: {
        duration: durations.normal,
        easing: easingCurves.standard,
        property: 'box-shadow'
      },
      all: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'all'
      }
    },
    
    // Input field transitions
    input: {
      border: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'border-color'
      },
      background: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'background-color'
      },
      focus: {
        duration: durations.normal,
        easing: easingCurves.decelerate,
        property: 'box-shadow, border-color'
      }
    },
    
    // Link transitions
    link: {
      color: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'color'
      },
      underline: {
        duration: durations.normal,
        easing: easingCurves.standard,
        property: 'text-decoration-color'
      }
    }
  },
  
  // Layout transitions
  layout: {
    // Content reveal/hide
    content: {
      fadeIn: {
        duration: durations.medium,
        easing: easingCurves.decelerate,
        property: 'opacity, transform'
      },
      fadeOut: {
        duration: durations.normal,
        easing: easingCurves.accelerate,
        property: 'opacity, transform'
      },
      slideDown: {
        duration: durations.medium,
        easing: easingCurves.decelerate,
        property: 'height, opacity'
      },
      slideUp: {
        duration: durations.normal,
        easing: easingCurves.accelerate,
        property: 'height, opacity'
      }
    },
    
    // Modal/Dialog transitions
    modal: {
      backdrop: {
        duration: durations.medium,
        easing: easingCurves.standard,
        property: 'opacity'
      },
      content: {
        duration: durations.medium,
        easing: easingCurves.emphasized,
        property: 'opacity, transform'
      }
    },
    
    // Sidebar/Navigation transitions
    navigation: {
      slide: {
        duration: durations.medium,
        easing: easingCurves.emphasized,
        property: 'transform'
      },
      fade: {
        duration: durations.normal,
        easing: easingCurves.standard,
        property: 'opacity'
      }
    }
  },
  
  // Feedback transitions
  feedback: {
    // Loading states
    loading: {
      spin: {
        duration: durations.slowest,
        easing: easingCurves.linear,
        property: 'transform',
        iterationCount: 'infinite'
      },
      pulse: {
        duration: durations.slower,
        easing: easingCurves.easeInOut,
        property: 'opacity',
        iterationCount: 'infinite',
        direction: 'alternate'
      },
      skeleton: {
        duration: durations.slower,
        easing: easingCurves.easeInOut,
        property: 'background-position',
        iterationCount: 'infinite'
      }
    },
    
    // Success/Error states
    status: {
      success: {
        duration: durations.medium,
        easing: easingCurves.bounce,
        property: 'transform, opacity'
      },
      error: {
        duration: durations.normal,
        easing: easingCurves.sharp,
        property: 'transform, background-color'
      },
      warning: {
        duration: durations.medium,
        easing: easingCurves.standard,
        property: 'background-color, border-color'
      }
    },
    
    // Toast/Notification transitions
    toast: {
      enter: {
        duration: durations.medium,
        easing: easingCurves.spring,
        property: 'opacity, transform'
      },
      exit: {
        duration: durations.normal,
        easing: easingCurves.accelerate,
        property: 'opacity, transform'
      }
    }
  },
  
  // Data visualization transitions
  data: {
    // Chart animations
    chart: {
      draw: {
        duration: durations.slower,
        easing: easingCurves.decelerate,
        property: 'stroke-dashoffset'
      },
      hover: {
        duration: durations.fast,
        easing: easingCurves.standard,
        property: 'transform, fill'
      }
    },
    
    // Progress indicators
    progress: {
      fill: {
        duration: durations.medium,
        easing: easingCurves.decelerate,
        property: 'width, background-color'
      },
      indeterminate: {
        duration: durations.slower,
        easing: easingCurves.linear,
        property: 'transform',
        iterationCount: 'infinite'
      }
    }
  }
} as const

// Pre-built transition combinations
export const transitionPresets = {
  // Common combinations
  fadeIn: `opacity ${durations.medium} ${easingCurves.decelerate}`,
  fadeOut: `opacity ${durations.normal} ${easingCurves.accelerate}`,
  
  slideInUp: `transform ${durations.medium} ${easingCurves.decelerate}, opacity ${durations.medium} ${easingCurves.decelerate}`,
  slideInDown: `transform ${durations.medium} ${easingCurves.decelerate}, opacity ${durations.medium} ${easingCurves.decelerate}`,
  slideInLeft: `transform ${durations.medium} ${easingCurves.decelerate}, opacity ${durations.medium} ${easingCurves.decelerate}`,
  slideInRight: `transform ${durations.medium} ${easingCurves.decelerate}, opacity ${durations.medium} ${easingCurves.decelerate}`,
  
  scaleIn: `transform ${durations.medium} ${easingCurves.spring}, opacity ${durations.medium} ${easingCurves.decelerate}`,
  scaleOut: `transform ${durations.normal} ${easingCurves.accelerate}, opacity ${durations.normal} ${easingCurves.accelerate}`,
  
  // Interactive presets
  buttonHover: `background-color ${durations.fast} ${easingCurves.standard}, transform ${durations.fast} ${easingCurves.standard}`,
  buttonPress: `transform ${durations.fast} ${easingCurves.sharp}`,
  
  inputFocus: `border-color ${durations.fast} ${easingCurves.standard}, box-shadow ${durations.normal} ${easingCurves.decelerate}`,
  
  // Layout presets
  modalEnter: `opacity ${durations.medium} ${easingCurves.standard}, transform ${durations.medium} ${easingCurves.emphasized}`,
  modalExit: `opacity ${durations.normal} ${easingCurves.accelerate}, transform ${durations.normal} ${easingCurves.accelerate}`,
  
  dropdownEnter: `opacity ${durations.normal} ${easingCurves.decelerate}, transform ${durations.normal} ${easingCurves.decelerate}`,
  dropdownExit: `opacity ${durations.fast} ${easingCurves.accelerate}, transform ${durations.fast} ${easingCurves.accelerate}`,
  
  // Loading presets
  spin: `transform ${durations.slowest} ${easingCurves.linear} infinite`,
  pulse: `opacity ${durations.slower} ${easingCurves.easeInOut} infinite alternate`,
  
  // All-purpose smooth transition
  smooth: `all ${durations.normal} ${easingCurves.standard}`,
  smoothFast: `all ${durations.fast} ${easingCurves.standard}`,
  smoothSlow: `all ${durations.medium} ${easingCurves.standard}`
} as const

// Animation keyframes (for CSS animations)
export const keyframes = {
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
    from: { transform: 'translateY(100%)', opacity: 0 },
    to: { transform: 'translateY(0)', opacity: 1 }
  },
  
  slideInDown: {
    from: { transform: 'translateY(-100%)', opacity: 0 },
    to: { transform: 'translateY(0)', opacity: 1 }
  },
  
  slideInLeft: {
    from: { transform: 'translateX(-100%)', opacity: 0 },
    to: { transform: 'translateX(0)', opacity: 1 }
  },
  
  slideInRight: {
    from: { transform: 'translateX(100%)', opacity: 0 },
    to: { transform: 'translateX(0)', opacity: 1 }
  },
  
  // Scale animations
  scaleIn: {
    from: { transform: 'scale(0.8)', opacity: 0 },
    to: { transform: 'scale(1)', opacity: 1 }
  },
  
  scaleOut: {
    from: { transform: 'scale(1)', opacity: 1 },
    to: { transform: 'scale(0.8)', opacity: 0 }
  },
  
  // Rotation animations
  spin: {
    from: { transform: 'rotate(0deg)' },
    to: { transform: 'rotate(360deg)' }
  },
  
  // Pulse animation
  pulse: {
    '0%': { opacity: 1 },
    '50%': { opacity: 0.5 },
    '100%': { opacity: 1 }
  },
  
  // Bounce animation
  bounce: {
    '0%, 20%, 53%, 80%, 100%': { transform: 'translate3d(0,0,0)' },
    '40%, 43%': { transform: 'translate3d(0, -30px, 0)' },
    '70%': { transform: 'translate3d(0, -15px, 0)' },
    '90%': { transform: 'translate3d(0, -4px, 0)' }
  },
  
  // Skeleton loading animation
  skeleton: {
    '0%': { backgroundPosition: '-200px 0' },
    '100%': { backgroundPosition: 'calc(200px + 100%) 0' }
  }
} as const

// Transition utility functions
export const transitionUtils = {
  // Create custom transition
  create: (
    property: string | string[],
    duration: keyof typeof durations = 'normal',
    easing: keyof typeof easingCurves = 'standard',
    delay = '0ms'
  ) => {
    const properties = Array.isArray(property) ? property : [property]
    return properties
      .map(prop => `${prop} ${durations[duration]} ${easingCurves[easing]} ${delay}`)
      .join(', ')
  },
  
  // Get preset transition
  getPreset: (preset: keyof typeof transitionPresets) => {
    return transitionPresets[preset]
  },
  
  // Get semantic transition
  getSemanticTransition: (
    category: keyof typeof semanticTransitions,
    component: string,
    variant: string
  ) => {
    const categoryObj = semanticTransitions[category] as any
    const componentObj = categoryObj?.[component] as any
    const transition = componentObj?.[variant]
    
    if (!transition) return transitionPresets.smooth
    
    return `${transition.property} ${transition.duration} ${transition.easing}`
  },
  
  // Generate CSS custom properties
  generateCSSVariables: () => {
    const cssVars: Record<string, string> = {}
    
    // Easing curves
    Object.entries(easingCurves).forEach(([key, value]) => {
      cssVars[`--easing-${key}`] = value
    })
    
    // Durations
    Object.entries(durations).forEach(([key, value]) => {
      cssVars[`--duration-${key}`] = value
    })
    
    // Transition presets
    Object.entries(transitionPresets).forEach(([key, value]) => {
      cssVars[`--transition-${key}`] = value
    })
    
    return cssVars
  },
  
  // Generate keyframe CSS
  generateKeyframes: () => {
    const keyframeCSS: Record<string, string> = {}
    
    Object.entries(keyframes).forEach(([name, frames]) => {
      const frameEntries = Object.entries(frames)
      const frameCSS = frameEntries
        .map(([key, styles]) => {
          const styleEntries = Object.entries(styles)
          const styleCSS = styleEntries
            .map(([prop, value]) => `${prop}: ${value}`)
            .join('; ')
          return `${key} { ${styleCSS} }`
        })
        .join(' ')
      
      keyframeCSS[name] = `@keyframes ${name} { ${frameCSS} }`
    })
    
    return keyframeCSS
  }
} as const

// Accessibility considerations for transitions
export const accessibilityTransitions = {
  // Respect prefers-reduced-motion
  reducedMotion: {
    // Safe transitions that don't cause vestibular issues
    safe: {
      opacity: `opacity ${durations.normal} ${easingCurves.linear}`,
      color: `color ${durations.normal} ${easingCurves.linear}`,
      backgroundColor: `background-color ${durations.normal} ${easingCurves.linear}`
    },
    
    // Disable problematic animations
    disabled: {
      transform: 'none',
      animationDuration: '0.01ms',
      animationIterationCount: '1',
      transitionDuration: '0.01ms'
    }
  },
  
  // Focus indicators should be immediate
  focus: {
    immediate: 'all 0ms linear',
    subtle: `box-shadow ${durations.fast} ${easingCurves.linear}`
  },
  
  // Essential transitions that improve accessibility
  essential: {
    // Loading indicators need animation for screen readers
    loading: `opacity ${durations.slower} ${easingCurves.easeInOut} infinite alternate`,
    
    // Status changes should be noticeable
    status: `background-color ${durations.medium} ${easingCurves.standard}`
  }
} as const

// Export types
export type EasingCurve = keyof typeof easingCurves
export type Duration = keyof typeof durations
export type TransitionPreset = keyof typeof transitionPresets
export type KeyframeName = keyof typeof keyframes

export default {
  easingCurves,
  durations,
  semanticTransitions,
  transitionPresets,
  keyframes,
  transitionUtils,
  accessibilityTransitions
}