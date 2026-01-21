import React from 'react';
import { useAppStore } from '../stores/appStore';
import { Div, Title, Text, Subhead, Card } from '@telegram-apps/ui-components';

const ProfilePage: React.FC = () => {
  const { user, tickets } = useAppStore();

  // Calculate statistics
  const totalTickets = tickets.length;
  const resolvedTickets = tickets.filter(ticket => ticket.status === 'решена' || ticket.status === 'закрыта').length;
  const inProgressTickets = tickets.filter(ticket => ticket.status === 'в работе').length;

  return (
    <Div style={{ paddingTop: '20px' }}>
      <Div style={{ textAlign: 'center', marginBottom: '24px' }}>
        {user?.photoUrl ? (
          <img 
            src={user.photoUrl} 
            alt="Аватар" 
            style={{ 
              width: '80px', 
              height: '80px', 
              borderRadius: '50%', 
              objectFit: 'cover',
              border: '2px solid var(--tg-theme-button-color, #0088cc)'
            }} 
          />
        ) : (
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '50%',
            backgroundColor: 'var(--tg-theme-button-color, #0088cc)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 16px',
            fontSize: '32px',
            color: 'white'
          }}>
            {user?.firstName.charAt(0).toUpperCase() || '?'}
          </div>
        )}
        <Title level="2" weight="1">
          {user?.firstName} {user?.lastName || ''}
        </Title>
        {user?.username && (
          <Subhead style={{ color: 'var(--tg-theme-hint-color)' }}>@{user.username}</Subhead>
        )}
        <Subhead style={{ marginTop: '4px', color: 'var(--tg-theme-hint-color)' }}>
          ID: {user?.telegramId}
        </Subhead>
      </Div>

      <Div>
        <Title level="3" weight="2" style={{ marginBottom: '12px' }}>Статистика обращений</Title>
        
        <Card style={{ padding: '16px', marginBottom: '12px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>Всего заявок</Text>
            <Text weight="1">{totalTickets}</Text>
          </div>
        </Card>
        
        <Card style={{ padding: '16px', marginBottom: '12px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>Решено</Text>
            <Text weight="1" style={{ color: '#388e3c' }}>{resolvedTickets}</Text>
          </div>
        </Card>
        
        <Card style={{ padding: '16px', marginBottom: '12px' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <Text>В работе</Text>
            <Text weight="1" style={{ color: '#ffa000' }}>{inProgressTickets}</Text>
          </div>
        </Card>
      </Div>

      <Div style={{ marginTop: '24px' }}>
        <Title level="3" weight="2" style={{ marginBottom: '12px' }}>Информация</Title>
        
        <Card style={{ padding: '16px' }}>
          <Subhead>Язык интерфейса</Subhead>
          <Text style={{ marginTop: '4px' }}>
            {user?.languageCode === 'ru' ? 'Русский' : user?.languageCode ? user.languageCode : 'English'}
          </Text>
          
          <Subhead style={{ marginTop: '12px' }}>Telegram WebApp</Subhead>
          <Text style={{ marginTop: '4px' }}>Версия: {window.Telegram?.WebApp?.version || 'не определена'}</Text>
          <Text style={{ marginTop: '4px' }}>Платформа: {window.Telegram?.WebApp?.platform || 'не определена'}</Text>
        </Card>
      </Div>
    </Div>
  );
};

export default ProfilePage;