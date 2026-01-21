import React from 'react';
import { useAppStore } from '../stores/appStore';
import { Button, Card, Div, Title, Text, Subhead, Link } from '@telegram-apps/ui-components';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

const TicketsPage: React.FC = () => {
  const { tickets, setSelectedTicket } = useAppStore();

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'новая':
        return { bg: '#e3f2fd', color: '#1976d2' };
      case 'в работе':
        return { bg: '#fff8e1', color: '#ffa000' };
      case 'решена':
      case 'закрыта':
        return { bg: '#e8f5e9', color: '#388e3c' };
      default:
        return { bg: '#f5f5f5', color: '#666666' };
    }
  };

  return (
    <Div style={{ paddingTop: '20px' }}>
      <Div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level="2" weight="1">Мои заявки</Title>
        <Button 
          size="s" 
          mode="outline"
          onClick={() => {
            // Navigate to Create Ticket page
            console.log('Navigate to create ticket');
          }}
        >
          Создать
        </Button>
      </Div>

      {tickets.length === 0 ? (
        <Div style={{ textAlign: 'center', marginTop: '40px' }}>
          <Text>У вас пока нет заявок</Text>
          <Div style={{ marginTop: '16px' }}>
            <Button 
              size="m" 
              mode="primary"
              onClick={() => {
                // Navigate to Create Ticket page
                console.log('Navigate to create ticket');
              }}
            >
              Создать первую заявку
            </Button>
          </Div>
        </Div>
      ) : (
        <Div style={{ marginTop: '16px' }}>
          {tickets.map(ticket => {
            const statusStyle = getStatusColor(ticket.status);
            return (
              <Card 
                key={ticket.id} 
                style={{ 
                  padding: '16px', 
                  marginTop: '12px',
                  cursor: 'pointer'
                }}
                onClick={() => setSelectedTicket(ticket)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <Text weight="1">{ticket.title}</Text>
                    <Subhead style={{ marginTop: '4px', color: 'var(--tg-theme-hint-color)' }}>
                      #{ticket.id.split('_')[1]} • {format(new Date(ticket.createdAt), 'dd MMM yyyy', { locale: ru })}
                    </Subhead>
                  </div>
                  <Div 
                    style={{ 
                      padding: '4px 10px', 
                      borderRadius: '12px',
                      fontSize: '12px',
                      backgroundColor: statusStyle.bg,
                      color: statusStyle.color
                    }}
                  >
                    {ticket.status}
                  </Div>
                </div>
              </Card>
            );
          })}
        </Div>
      )}
    </Div>
  );
};

export default TicketsPage;