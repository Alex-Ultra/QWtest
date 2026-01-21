import React, { useState, useRef, useEffect } from 'react';
import { useAppStore } from '../stores/appStore';
import { Button, Div, Text, Subhead, Input } from '@telegram-apps/ui-components';

const ChatPage: React.FC = () => {
  const { selectedTicket, getMessagesByTicket, addMessage } = useAppStore();
  const [message, setMessage] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Get messages for the selected ticket
  const ticketMessages = selectedTicket ? getMessagesByTicket(selectedTicket.id) : [];

  const handleSendMessage = () => {
    if (message.trim() && selectedTicket) {
      addMessage({
        ticketId: selectedTicket.id,
        senderId: 1, // Current user ID
        senderType: 'user',
        content: message
      });
      setMessage('');
    }
  };

  // Scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom();
  }, [ticketMessages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  if (!selectedTicket) {
    return (
      <Div style={{ paddingTop: '20px', textAlign: 'center' }}>
        <Text>Выберите заявку, чтобы начать чат</Text>
      </Div>
    );
  }

  return (
    <Div style={{ 
      paddingTop: '20px', 
      height: 'calc(100vh - 120px)', 
      display: 'flex', 
      flexDirection: 'column' 
    }}>
      <Div style={{ marginBottom: '16px' }}>
        <Text weight="1">Чат по заявке #{selectedTicket.id.split('_')[1]}</Text>
        <Subhead style={{ marginTop: '4px', color: 'var(--tg-theme-hint-color)' }}>
          {selectedTicket.title}
        </Subhead>
      </Div>

      <Div style={{ 
        flex: 1, 
        overflowY: 'auto', 
        maxHeight: '70vh', 
        border: '1px solid var(--tg-theme-hint-color, #ccc)', 
        borderRadius: '12px', 
        padding: '12px',
        marginBottom: '12px'
      }}>
        {ticketMessages.length === 0 ? (
          <Div style={{ textAlign: 'center', padding: '20px' }}>
            <Text>Пока нет сообщений</Text>
          </Div>
        ) : (
          ticketMessages.map((msg, index) => (
            <Div 
              key={msg.id} 
              style={{ 
                marginBottom: '12px', 
                textAlign: msg.senderType === 'user' ? 'right' : 'left' 
              }}
            >
              <Div
                style={{
                  display: 'inline-block',
                  padding: '8px 12px',
                  borderRadius: '18px',
                  backgroundColor: msg.senderType === 'user' 
                    ? 'var(--tg-theme-button-color, #0088cc)' 
                    : 'var(--tg-theme-hint-color, #e0e0e0)',
                  color: msg.senderType === 'user' ? 'white' : 'black',
                  maxWidth: '80%'
                }}
              >
                <Text style={{ wordBreak: 'break-word' }}>{msg.content}</Text>
              </Div>
              <Subhead 
                style={{ 
                  marginTop: '4px', 
                  fontSize: '12px',
                  color: 'var(--tg-theme-hint-color, #888)'
                }}
              >
                {/* Format date */}
              </Subhead>
            </Div>
          ))
        )}
        <div ref={messagesEndRef} />
      </Div>

      <Div style={{ 
        display: 'flex', 
        gap: '8px',
        alignItems: 'center'
      }}>
        <Input 
          placeholder="Введите сообщение..." 
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          style={{ flex: 1 }}
          onKeyPress={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              handleSendMessage();
            }
          }}
        />
        <Button 
          size="m" 
          mode="primary"
          disabled={!message.trim()}
          onClick={handleSendMessage}
        >
          Отправить
        </Button>
      </Div>
    </Div>
  );
};

export default ChatPage;