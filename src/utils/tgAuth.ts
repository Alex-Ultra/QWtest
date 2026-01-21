declare global {
  interface Window {
    Telegram?: {
      WebApp?: {
        initDataUnsafe?: any;
        initData?: string;
        colorScheme?: string;
        themeParams?: Record<string, string>;
        version?: string;
        platform?: string;
        isExpanded?: boolean;
        viewportHeight?: number;
        viewportStableHeight?: number;
        headerColor?: string;
        backgroundColor?: string;
        BackButton?: {
          show: () => void;
          hide: () => void;
          onClick: (callback: () => void) => void;
          offClick: (callback: () => void) => void;
        };
        MainButton?: {
          text: string;
          color: string;
          textColor: string;
          isVisible: boolean;
          isActive: boolean;
          isProgressVisible: boolean;
          setText: (text: string) => void;
          onClick: (callback: () => void) => void;
          offClick: (callback: () => void) => void;
          show: () => void;
          hide: () => void;
          enable: () => void;
          disable: () => void;
          showProgress: (leaveActive: boolean) => void;
          hideProgress: () => void;
        };
        ready: () => void;
        expand: () => void;
      };
    };
  }
}

export const initTgUser = () => {
  if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
    const tgWebApp = window.Telegram.WebApp;
    tgWebApp.ready();
    
    const userData = tgWebApp.initDataUnsafe?.user;
    
    if (userData) {
      return {
        id: userData.id,
        telegramId: userData.id.toString(),
        firstName: userData.first_name || '',
        lastName: userData.last_name,
        username: userData.username,
        languageCode: userData.language_code,
        photoUrl: userData.photo_url,
      };
    }
  }
  
  // Return mock user for development
  return {
    id: 1,
    telegramId: '123456789',
    firstName: 'Иван',
    lastName: 'Иванов',
    username: 'ivan_ivanov',
    languageCode: 'ru',
    photoUrl: undefined,
  };
};

export const setupTgTheme = () => {
  if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
    const tgWebApp = window.Telegram.WebApp;
    
    // Apply Telegram theme colors to CSS variables
    const updateTheme = () => {
      const themeParams = tgWebApp.themeParams || {};
      
      document.documentElement.style.setProperty(
        '--tg-theme-bg-color', 
        themeParams.bg_color || '#ffffff'
      );
      document.documentElement.style.setProperty(
        '--tg-theme-text-color', 
        themeParams.text_color || '#000000'
      );
      document.documentElement.style.setProperty(
        '--tg-theme-hint-color', 
        themeParams.hint_color || '#888888'
      );
      document.documentElement.style.setProperty(
        '--tg-theme-link-color', 
        themeParams.link_color || '#2678b6'
      );
      document.documentElement.style.setProperty(
        '--tg-theme-button-color', 
        themeParams.button_color || '#3390ec'
      );
      document.documentElement.style.setProperty(
        '--tg-theme-button-text-color', 
        themeParams.button_text_color || '#ffffff'
      );
    };
    
    updateTheme();
    tgWebApp.onEvent('themeChanged', updateTheme);
  }
};