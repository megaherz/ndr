/**
 * Telegram Mini Apps SDK Integration
 * 
 * This module provides a wrapper around the Telegram WebApp API
 * for seamless integration with the Nitro Drag Royale Mini App.
 */

// Extend the Window interface to include Telegram WebApp
declare global {
  interface Window {
    Telegram: {
      WebApp: TelegramWebApp;
    };
  }
}

// Telegram WebApp API types
interface TelegramWebApp {
  initData: string;
  initDataUnsafe: TelegramWebAppInitData;
  version: string;
  platform: string;
  colorScheme: 'light' | 'dark';
  themeParams: TelegramThemeParams;
  isExpanded: boolean;
  viewportHeight: number;
  viewportStableHeight: number;
  isClosingConfirmationEnabled: boolean;
  headerColor: string;
  backgroundColor: string;
  BackButton: TelegramBackButton;
  MainButton: TelegramMainButton;
  HapticFeedback: TelegramHapticFeedback;
  ready(): void;
  expand(): void;
  close(): void;
  enableClosingConfirmation(): void;
  disableClosingConfirmation(): void;
  onEvent(eventType: string, eventHandler: () => void): void;
  offEvent(eventType: string, eventHandler: () => void): void;
  sendData(data: string): void;
  openLink(url: string): void;
  openTelegramLink(url: string): void;
  showPopup(params: TelegramPopupParams, callback?: (buttonId: string) => void): void;
  showAlert(message: string, callback?: () => void): void;
  showConfirm(message: string, callback?: (confirmed: boolean) => void): void;
  showScanQrPopup(params: TelegramScanQrPopupParams, callback?: (text: string) => void): void;
  closeScanQrPopup(): void;
  readTextFromClipboard(callback?: (text: string) => void): void;
  requestWriteAccess(callback?: (granted: boolean) => void): void;
  requestContact(callback?: (granted: boolean, contact?: TelegramContact) => void): void;
}

interface TelegramWebAppInitData {
  query_id?: string;
  user?: TelegramUser;
  receiver?: TelegramUser;
  chat?: TelegramChat;
  chat_type?: string;
  chat_instance?: string;
  start_param?: string;
  can_send_after?: number;
  auth_date: number;
  hash: string;
}

interface TelegramUser {
  id: number;
  is_bot?: boolean;
  first_name: string;
  last_name?: string;
  username?: string;
  language_code?: string;
  is_premium?: boolean;
  added_to_attachment_menu?: boolean;
  allows_write_to_pm?: boolean;
  photo_url?: string;
}

interface TelegramChat {
  id: number;
  type: 'group' | 'supergroup' | 'channel';
  title: string;
  username?: string;
  photo_url?: string;
}

interface TelegramThemeParams {
  bg_color?: string;
  text_color?: string;
  hint_color?: string;
  link_color?: string;
  button_color?: string;
  button_text_color?: string;
  secondary_bg_color?: string;
}

interface TelegramBackButton {
  isVisible: boolean;
  onClick(callback: () => void): void;
  offClick(callback: () => void): void;
  show(): void;
  hide(): void;
}

interface TelegramMainButton {
  text: string;
  color: string;
  textColor: string;
  isVisible: boolean;
  isActive: boolean;
  isProgressVisible: boolean;
  setText(text: string): void;
  onClick(callback: () => void): void;
  offClick(callback: () => void): void;
  show(): void;
  hide(): void;
  enable(): void;
  disable(): void;
  showProgress(leaveActive?: boolean): void;
  hideProgress(): void;
  setParams(params: TelegramMainButtonParams): void;
}

interface TelegramMainButtonParams {
  text?: string;
  color?: string;
  text_color?: string;
  is_active?: boolean;
  is_visible?: boolean;
}

interface TelegramHapticFeedback {
  impactOccurred(style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'): void;
  notificationOccurred(type: 'error' | 'success' | 'warning'): void;
  selectionChanged(): void;
}

interface TelegramPopupParams {
  title?: string;
  message: string;
  buttons?: TelegramPopupButton[];
}

interface TelegramPopupButton {
  id?: string;
  type?: 'default' | 'ok' | 'close' | 'cancel' | 'destructive';
  text?: string;
}

interface TelegramScanQrPopupParams {
  text?: string;
}

interface TelegramContact {
  contact: {
    user_id: number;
    phone_number: string;
    first_name: string;
    last_name?: string;
    vcard?: string;
  };
}

/**
 * Telegram Mini Apps SDK wrapper class
 */
export class TelegramSDK {
  private static instance: TelegramSDK;
  private webApp: TelegramWebApp | null = null;
  private isReady = false;

  private constructor() {
    this.initialize();
  }

  /**
   * Get singleton instance of TelegramSDK
   */
  public static getInstance(): TelegramSDK {
    if (!TelegramSDK.instance) {
      TelegramSDK.instance = new TelegramSDK();
    }
    return TelegramSDK.instance;
  }

  /**
   * Initialize the Telegram WebApp
   */
  private initialize(): void {
    if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
      this.webApp = window.Telegram.WebApp;
      this.webApp.ready();
      this.isReady = true;
      
      // Configure the app
      this.webApp.expand();
      this.webApp.enableClosingConfirmation();
      
      console.log('Telegram WebApp initialized:', {
        version: this.webApp.version,
        platform: this.webApp.platform,
        colorScheme: this.webApp.colorScheme,
        user: this.webApp.initDataUnsafe.user,
      });
    } else {
      console.warn('Telegram WebApp not available. Running in development mode.');
    }
  }

  /**
   * Check if running inside Telegram
   */
  public isTelegramEnvironment(): boolean {
    return this.isReady && this.webApp !== null;
  }

  /**
   * Get the current user information
   */
  public getUser(): TelegramUser | null {
    return this.webApp?.initDataUnsafe.user || null;
  }

  /**
   * Get the initialization data for backend authentication
   */
  public getInitData(): string {
    return this.webApp?.initData || '';
  }

  /**
   * Get the theme parameters
   */
  public getThemeParams(): TelegramThemeParams {
    return this.webApp?.themeParams || {};
  }

  /**
   * Get the current color scheme
   */
  public getColorScheme(): 'light' | 'dark' {
    return this.webApp?.colorScheme || 'light';
  }

  /**
   * Show haptic feedback
   */
  public hapticFeedback(type: 'impact' | 'notification' | 'selection', style?: string): void {
    if (!this.webApp?.HapticFeedback) return;

    switch (type) {
      case 'impact':
        this.webApp.HapticFeedback.impactOccurred((style as 'light' | 'medium' | 'heavy') || 'medium');
        break;
      case 'notification':
        this.webApp.HapticFeedback.notificationOccurred((style as 'error' | 'success' | 'warning') || 'success');
        break;
      case 'selection':
        this.webApp.HapticFeedback.selectionChanged();
        break;
    }
  }

  /**
   * Show the main button with custom text and callback
   */
  public showMainButton(text: string, callback: () => void): void {
    if (!this.webApp?.MainButton) return;

    this.webApp.MainButton.setText(text);
    this.webApp.MainButton.onClick(callback);
    this.webApp.MainButton.show();
  }

  /**
   * Hide the main button
   */
  public hideMainButton(): void {
    if (!this.webApp?.MainButton) return;
    this.webApp.MainButton.hide();
  }

  /**
   * Show the back button with callback
   */
  public showBackButton(callback: () => void): void {
    if (!this.webApp?.BackButton) return;

    this.webApp.BackButton.onClick(callback);
    this.webApp.BackButton.show();
  }

  /**
   * Hide the back button
   */
  public hideBackButton(): void {
    if (!this.webApp?.BackButton) return;
    this.webApp.BackButton.hide();
  }

  /**
   * Show an alert dialog
   */
  public showAlert(message: string): Promise<void> {
    return new Promise((resolve) => {
      if (this.webApp?.showAlert) {
        this.webApp.showAlert(message, () => resolve());
      } else {
        alert(message);
        resolve();
      }
    });
  }

  /**
   * Show a confirmation dialog
   */
  public showConfirm(message: string): Promise<boolean> {
    return new Promise((resolve) => {
      if (this.webApp?.showConfirm) {
        this.webApp.showConfirm(message, (confirmed) => resolve(confirmed));
      } else {
        resolve(confirm(message));
      }
    });
  }

  /**
   * Close the Mini App
   */
  public close(): void {
    if (this.webApp?.close) {
      this.webApp.close();
    } else {
      window.close();
    }
  }

  /**
   * Open a link in the browser
   */
  public openLink(url: string): void {
    if (this.webApp?.openLink) {
      this.webApp.openLink(url);
    } else {
      window.open(url, '_blank');
    }
  }

  /**
   * Get viewport information
   */
  public getViewport(): { height: number; stableHeight: number; isExpanded: boolean } {
    return {
      height: this.webApp?.viewportHeight || window.innerHeight,
      stableHeight: this.webApp?.viewportStableHeight || window.innerHeight,
      isExpanded: this.webApp?.isExpanded || false,
    };
  }

  /**
   * Listen for viewport changes
   */
  public onViewportChanged(callback: () => void): void {
    if (this.webApp?.onEvent) {
      this.webApp.onEvent('viewportChanged', callback);
    } else {
      window.addEventListener('resize', callback);
    }
  }

  /**
   * Remove viewport change listener
   */
  public offViewportChanged(callback: () => void): void {
    if (this.webApp?.offEvent) {
      this.webApp.offEvent('viewportChanged', callback);
    } else {
      window.removeEventListener('resize', callback);
    }
  }

  /**
   * Listen for theme changes
   */
  public onThemeChanged(callback: () => void): void {
    if (this.webApp?.onEvent) {
      this.webApp.onEvent('themeChanged', callback);
    }
  }

  /**
   * Remove theme change listener
   */
  public offThemeChanged(callback: () => void): void {
    if (this.webApp?.offEvent) {
      this.webApp.offEvent('themeChanged', callback);
    }
  }
}

// Export singleton instance
export const telegramSDK = TelegramSDK.getInstance();