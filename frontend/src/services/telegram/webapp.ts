/**
 * Telegram WebApp API utilities and helpers
 */

import { telegramSDK } from './sdk';

/**
 * Extract and validate Telegram initData for backend authentication
 */
interface TelegramUser {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  language_code?: string;
  is_premium?: boolean;
}

export function extractInitData(): {
  initData: string;
  user: TelegramUser | null;
  isValid: boolean;
} {
  const initData = telegramSDK.getInitData();
  const user = telegramSDK.getUser();
  
  return {
    initData,
    user,
    isValid: Boolean(initData && user),
  };
}

/**
 * Apply Telegram theme to CSS custom properties
 */
export function applyTelegramTheme(): void {
  const themeParams = telegramSDK.getThemeParams();
  const colorScheme = telegramSDK.getColorScheme();
  
  // Set CSS custom properties based on Telegram theme
  const root = document.documentElement;
  
  if (themeParams.bg_color) {
    root.style.setProperty('--tg-bg-color', themeParams.bg_color);
  }
  
  if (themeParams.text_color) {
    root.style.setProperty('--tg-text-color', themeParams.text_color);
  }
  
  if (themeParams.hint_color) {
    root.style.setProperty('--tg-hint-color', themeParams.hint_color);
  }
  
  if (themeParams.link_color) {
    root.style.setProperty('--tg-link-color', themeParams.link_color);
  }
  
  if (themeParams.button_color) {
    root.style.setProperty('--tg-button-color', themeParams.button_color);
  }
  
  if (themeParams.button_text_color) {
    root.style.setProperty('--tg-button-text-color', themeParams.button_text_color);
  }
  
  if (themeParams.secondary_bg_color) {
    root.style.setProperty('--tg-secondary-bg-color', themeParams.secondary_bg_color);
  }
  
  // Set color scheme
  root.style.setProperty('--tg-color-scheme', colorScheme);
  root.setAttribute('data-theme', colorScheme);
}

/**
 * Initialize Telegram WebApp with proper configuration
 */
export function initializeTelegramWebApp(): void {
  if (!telegramSDK.isTelegramEnvironment()) {
    console.log('Not running in Telegram environment, skipping WebApp initialization');
    return;
  }
  
  // Apply theme
  applyTelegramTheme();
  
  // Listen for theme changes
  telegramSDK.onThemeChanged(() => {
    applyTelegramTheme();
  });
  
  // Handle viewport changes for responsive design
  telegramSDK.onViewportChanged(() => {
    const viewport = telegramSDK.getViewport();
    document.documentElement.style.setProperty('--tg-viewport-height', `${viewport.height}px`);
    document.documentElement.style.setProperty('--tg-viewport-stable-height', `${viewport.stableHeight}px`);
  });
  
  // Set initial viewport
  const viewport = telegramSDK.getViewport();
  document.documentElement.style.setProperty('--tg-viewport-height', `${viewport.height}px`);
  document.documentElement.style.setProperty('--tg-viewport-stable-height', `${viewport.stableHeight}px`);
}

/**
 * Show success feedback with haptic
 */
export function showSuccess(message?: string): void {
  telegramSDK.hapticFeedback('notification', 'success');
  if (message) {
    telegramSDK.showAlert(message);
  }
}

/**
 * Show error feedback with haptic
 */
export function showError(message: string): void {
  telegramSDK.hapticFeedback('notification', 'error');
  telegramSDK.showAlert(message);
}

/**
 * Show warning feedback with haptic
 */
export function showWarning(message: string): void {
  telegramSDK.hapticFeedback('notification', 'warning');
  telegramSDK.showAlert(message);
}

/**
 * Provide tactile feedback for button presses
 */
export function buttonPress(style: 'light' | 'medium' | 'heavy' = 'medium'): void {
  telegramSDK.hapticFeedback('impact', style);
}

/**
 * Provide tactile feedback for selection changes
 */
export function selectionChanged(): void {
  telegramSDK.hapticFeedback('selection');
}

/**
 * Get user's preferred language from Telegram
 */
export function getUserLanguage(): string {
  const user = telegramSDK.getUser();
  return user?.language_code || 'en';
}

/**
 * Check if user is premium
 */
export function isUserPremium(): boolean {
  const user = telegramSDK.getUser();
  return user?.is_premium || false;
}

/**
 * Get user display name
 */
export function getUserDisplayName(): string {
  const user = telegramSDK.getUser();
  if (!user) return 'Player';
  
  const parts = [user.first_name, user.last_name].filter(Boolean);
  return parts.join(' ') || user.username || 'Player';
}

/**
 * Check if running in development mode
 */
export function isDevelopmentMode(): boolean {
  return !telegramSDK.isTelegramEnvironment() || import.meta.env.DEV;
}

/**
 * Safe wrapper for async operations with error handling
 */
export async function safeAsync<T>(
  operation: () => Promise<T>,
  fallback: T,
  errorMessage?: string
): Promise<T> {
  try {
    return await operation();
  } catch (error) {
    console.error('Async operation failed:', error);
    if (errorMessage) {
      showError(errorMessage);
    }
    return fallback;
  }
}