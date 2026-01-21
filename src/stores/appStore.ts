import { create } from 'zustand';
import { User, Ticket, Message, TicketStatus, TicketCategory } from '../types';

interface AppState {
  user: User | null;
  tickets: Ticket[];
  messages: Message[];
  selectedTicket: Ticket | null;
  setUser: (user: User) => void;
  addTicket: (ticket: Omit<Ticket, 'id' | 'createdAt' | 'updatedAt'>) => void;
  updateTicketStatus: (ticketId: string, status: TicketStatus) => void;
  setSelectedTicket: (ticket: Ticket | null) => void;
  addMessage: (message: Omit<Message, 'id' | 'timestamp'>) => void;
  getMessagesByTicket: (ticketId: string) => Message[];
}

export const useAppStore = create<AppState>((set, get) => ({
  user: null,
  tickets: [],
  messages: [],
  selectedTicket: null,
  
  setUser: (user) => set({ user }),
  
  addTicket: (ticketData) => {
    const newTicket: Ticket = {
      ...ticketData,
      id: `ticket_${Date.now()}`,
      createdAt: new Date(),
      updatedAt: new Date(),
    };
    
    set((state) => ({ 
      tickets: [...state.tickets, newTicket],
      selectedTicket: newTicket
    }));
  },
  
  updateTicketStatus: (ticketId, status) => {
    set((state) => ({
      tickets: state.tickets.map(ticket => 
        ticket.id === ticketId 
          ? { ...ticket, status, updatedAt: new Date() } 
          : ticket
      )
    }));
  },
  
  setSelectedTicket: (ticket) => set({ selectedTicket: ticket }),
  
  addMessage: (messageData) => {
    const newMessage: Message = {
      ...messageData,
      id: `msg_${Date.now()}`,
      timestamp: new Date()
    };
    
    set((state) => ({ 
      messages: [...state.messages, newMessage]
    }));
  },
  
  getMessagesByTicket: (ticketId) => {
    return get().messages.filter(msg => msg.ticketId === ticketId);
  }
}));