import React, { useState, useEffect } from 'react';
import { useAppStore } from '../../stores/appStore';
import { setupTgTheme } from '../../utils/tgAuth';
import WelcomePage from '../../pages/WelcomePage';
import TicketsPage from '../../pages/TicketsPage';
import ChatPage from '../../pages/ChatPage';
import ProfilePage from '../../pages/ProfilePage';
import { Cell, Div, Footer, Header, Navigation, NavTab, Tabbar, TabbarItem } from '@telegram-apps/ui-components';

const MainLayout: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'home' | 'tickets' | 'chat' | 'profile'>('home');
  const { user } = useAppStore();

  useEffect(() => {
    setupTgTheme();
  }, []);

  const renderContent = () => {
    switch (activeTab) {
      case 'home':
        return <WelcomePage />;
      case 'tickets':
        return <TicketsPage />;
      case 'chat':
        return <ChatPage />;
      case 'profile':
        return <ProfilePage />;
      default:
        return <WelcomePage />;
    }
  };

  return (
    <div className="app-container">
      <Div>{renderContent()}</Div>
      
      <Footer style={{ height: '60px' }} />
      
      <Tabbar 
        style={{ 
          position: 'fixed', 
          bottom: 0, 
          left: 0, 
          right: 0,
          backgroundColor: 'var(--tg-theme-bg-color, white)',
          borderTop: '1px solid var(--tg-theme-hint-color, #ccc)'
        }}
      >
        <TabbarItem
          icon={
            <Cell
              media={
                <div style={{
                  width: '24px',
                  height: '24px',
                  backgroundColor: activeTab === 'home' ? 'var(--tg-theme-button-color, #0088cc)' : 'var(--tg-theme-hint-color, #888)',
                  borderRadius: '50%',
                }}
                />
              }
            />
          }
          text="Главная"
          selected={activeTab === 'home'}
          onClick={() => setActiveTab('home')}
        />
        <TabbarItem
          icon={
            <Cell
              media={
                <div style={{
                  width: '24px',
                  height: '24px',
                  backgroundColor: activeTab === 'tickets' ? 'var(--tg-theme-button-color, #0088cc)' : 'var(--tg-theme-hint-color, #888)',
                  borderRadius: '50%',
                }}
                />
              }
            />
          }
          text="Заявки"
          selected={activeTab === 'tickets'}
          onClick={() => setActiveTab('tickets')}
        />
        <TabbarItem
          icon={
            <Cell
              media={
                <div style={{
                  width: '24px',
                  height: '24px',
                  backgroundColor: activeTab === 'chat' ? 'var(--tg-theme-button-color, #0088cc)' : 'var(--tg-theme-hint-color, #888)',
                  borderRadius: '50%',
                }}
                />
              }
            />
          }
          text="Чат"
          selected={activeTab === 'chat'}
          onClick={() => setActiveTab('chat')}
        />
        <TabbarItem
          icon={
            <Cell
              media={
                <div style={{
                  width: '24px',
                  height: '24px',
                  backgroundColor: activeTab === 'profile' ? 'var(--tg-theme-button-color, #0088cc)' : 'var(--tg-theme-hint-color, #888)',
                  borderRadius: '50%',
                }}
                />
              }
            />
          }
          text="Профиль"
          selected={activeTab === 'profile'}
          onClick={() => setActiveTab('profile')}
        />
      </Tabbar>
    </div>
  );
};

export default MainLayout;