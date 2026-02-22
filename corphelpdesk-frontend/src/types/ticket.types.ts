export type TicketStatus = 'new' | 'open' | 'pending' | 'resolved' | 'closed';
export type TicketPriority = 'low' | 'normal' | 'high' | 'critical';

export interface Ticket {
  id: string;
  userId: string;
  agentId?: string;
  subject: string;
  description: string;
  status: TicketStatus;
  priority: TicketPriority;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTicketRequest {
  subject: string;
  description: string;
  priority: TicketPriority;
}

export interface UpdateTicketStatusRequest {
  status: TicketStatus;
}

export interface Message {
  id: string;
  ticketId: string;
  userId: string;
  content: string;
  attachments?: string[];
  isInternal: boolean;
  createdAt: string;
}

export interface CreateMessageRequest {
  content: string;
  attachments?: string[];
  isInternal?: boolean;
}