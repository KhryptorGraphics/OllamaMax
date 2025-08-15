/**
 * Metric Card Component - React Native
 * 
 * Native metric card with animations and touch interactions.
 */

import React, { useRef, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Animated,
  Platform,
  ViewStyle,
} from 'react-native';
import LinearGradient from 'react-native-linear-gradient';
import { useTheme } from '../contexts/ThemeContext';
import { colorUtils } from '../theme/colors';

interface MetricCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  trend?: number;
  icon: string;
  color: string;
  onPress?: () => void;
  style?: ViewStyle;
  animated?: boolean;
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  subtitle,
  trend,
  icon,
  color,
  onPress,
  style,
  animated = true,
}) => {
  const { theme } = useTheme();
  const scaleAnim = useRef(new Animated.Value(1)).current;
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(50)).current;

  // Animation on mount
  useEffect(() => {
    if (animated) {
      Animated.parallel([
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 600,
          useNativeDriver: true,
        }),
        Animated.spring(slideAnim, {
          toValue: 0,
          tension: 100,
          friction: 8,
          useNativeDriver: true,
        }),
      ]).start();
    } else {
      fadeAnim.setValue(1);
      slideAnim.setValue(0);
    }
  }, [animated, fadeAnim, slideAnim]);

  // Press animations
  const handlePressIn = () => {
    Animated.spring(scaleAnim, {
      toValue: 0.95,
      useNativeDriver: true,
    }).start();
  };

  const handlePressOut = () => {
    Animated.spring(scaleAnim, {
      toValue: 1,
      useNativeDriver: true,
    }).start();
  };

  // Get trend color and icon
  const getTrendInfo = () => {
    if (trend === undefined) return null;
    
    const isPositive = trend > 0;
    const isNegative = trend < 0;
    
    return {
      color: isPositive ? theme.colors.success : isNegative ? theme.colors.error : theme.colors.textSecondary,
      icon: isPositive ? '↗' : isNegative ? '↘' : '→',
      text: `${isPositive ? '+' : ''}${trend}%`,
    };
  };

  const trendInfo = getTrendInfo();
  const styles = createStyles(theme, color);

  const cardContent = (
    <Animated.View
      style={[
        styles.container,
        style,
        {
          opacity: fadeAnim,
          transform: [
            { scale: scaleAnim },
            { translateY: slideAnim },
          ],
        },
      ]}
    >
      {/* Background gradient */}
      <LinearGradient
        colors={[
          colorUtils.withOpacity(color, 0.1),
          colorUtils.withOpacity(color, 0.05),
        ]}
        style={styles.background}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      />

      {/* Header */}
      <View style={styles.header}>
        <View style={[styles.iconContainer, { backgroundColor: colorUtils.withOpacity(color, 0.15) }]}>
          <Text style={styles.icon}>{icon}</Text>
        </View>
        
        {trendInfo && (
          <View style={[styles.trendContainer, { backgroundColor: colorUtils.withOpacity(trendInfo.color, 0.1) }]}>
            <Text style={[styles.trendIcon, { color: trendInfo.color }]}>
              {trendInfo.icon}
            </Text>
            <Text style={[styles.trendText, { color: trendInfo.color }]}>
              {Math.abs(trend!)}%
            </Text>
          </View>
        )}
      </View>

      {/* Content */}
      <View style={styles.content}>
        <Text style={styles.value} numberOfLines={1} adjustsFontSizeToFit>
          {value}
        </Text>
        <Text style={styles.title} numberOfLines={2}>
          {title}
        </Text>
        {subtitle && (
          <Text style={styles.subtitle} numberOfLines={1}>
            {subtitle}
          </Text>
        )}
      </View>

      {/* Accent line */}
      <View style={[styles.accentLine, { backgroundColor: color }]} />
    </Animated.View>
  );

  if (onPress) {
    return (
      <TouchableOpacity
        onPress={onPress}
        onPressIn={handlePressIn}
        onPressOut={handlePressOut}
        activeOpacity={0.9}
        style={styles.touchable}
      >
        {cardContent}
      </TouchableOpacity>
    );
  }

  return cardContent;
};

const createStyles = (theme: any, color: string) => StyleSheet.create({
  touchable: {
    borderRadius: 16,
  },
  container: {
    backgroundColor: theme.colors.surface,
    borderRadius: 16,
    padding: 16,
    minHeight: 120,
    position: 'relative',
    overflow: 'hidden',
    ...Platform.select({
      ios: {
        shadowColor: theme.colors.shadow,
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 8,
      },
      android: {
        elevation: 4,
      },
    }),
  },
  background: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  iconContainer: {
    width: 40,
    height: 40,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  icon: {
    fontSize: 20,
  },
  trendContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  trendIcon: {
    fontSize: 12,
    marginRight: 2,
  },
  trendText: {
    fontSize: 12,
    fontWeight: '600',
  },
  content: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  value: {
    fontSize: 24,
    fontWeight: 'bold',
    color: theme.colors.text,
    marginBottom: 4,
  },
  title: {
    fontSize: 14,
    fontWeight: '500',
    color: theme.colors.textSecondary,
    lineHeight: 18,
  },
  subtitle: {
    fontSize: 12,
    color: theme.colors.textMuted,
    marginTop: 2,
  },
  accentLine: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: 3,
  },
});

export default MetricCard;
