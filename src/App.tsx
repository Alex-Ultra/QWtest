import React, { useEffect } from 'react';
import { initTgUser } from './utils/tgAuth';
import MainLayout from './components/layout/MainLayout';
import { useAppStore } from './stores/appStore';
import './App.css';

const App: React.FC = () => {
  const { setUser, user } = useAppStore();

  useEffect(() => {
    // Initialize user from Telegram WebApp
    const tgUser = initTgUser();
    if (tgUser) {
      setUser(tgUser);
    }
  }, [setUser]);

  if (!user) {
    return <div>Загрузка...</div>;
  }

  return (
    <MainLayout />
  );
};

export default App;