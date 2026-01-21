import React from 'react';
import { useAppStore } from '../stores/appStore';
import { Button, Card, Div, Title, Text, Subhead } from '@telegram-apps/ui-components';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

const WelcomePage: React.FC = () => {
  const { user, tickets } = useAppStore();

  const recentTickets = tickets.slice(0, 3);

  return (
    <Div style={{ paddingTop: '20px' }}>
      <Title level="1" weight="1">Добро пожаловать!</Title>
      <Text style={{ marginTop: '8px' }}>
        {user && `Привет, ${user.firstName}!`}
      </Text>
      
      <Div style={{ marginTop: '24px' }}>
        <Button 
          size="l" 
          mode="primary"
          onClick={() => {
            // Navigate to Create Ticket page
            console.log('Navigate to create ticket');
          }}
        >
          Создать заявку
        </Button>
      </Div>

      {recentTickets.length > 0 && (
        <Div style={{ marginTop: '24px' }}>
          <Subhead weight="1">Последние обращения</Subhead>
          {recentTickets.map(ticket => (
            <Card key={ticket.id} style={{ padding: '12px', marginTop: '8px' }}>
              <Text weight="1">{ticket.title}</Text>
              <Subhead style={{ marginTop: '4px', color: 'var(--tg-theme-hint-color)' }}>
                {format(new Date(ticket.createdAt), 'dd MMMM yyyy', { locale: ru })}
              </Subhead>
              <Div style={{ 
                marginTop: '8px', 
                padding: '4px 8px', 
                borderRadius: '8px',
                display: 'inline-block',
                fontSize: '12px',
                backgroundColor: 
                  ticket.status === 'новая' ? '#e3f2fd' : 
                  ticket.status === 'в работе' ? '#fff8e1' : 
                  '#e8f5e9',
                color: 
                  ticket.status === 'новая' ? '#1976d2' : 
                  ticket.status === 'в работе' ? '#ffa000' : 
                  '#388e3c'
              }}>
                {ticket.status}
              </Div>
            </Card>
          ))}
        </Div>
      )}
    </Div>
  );
};

export default WelcomePage;