export interface User {
  id: number;
  telegramId: string;
  firstName: string;
  lastName?: string;
  username?: string;
  languageCode?: string;
  photoUrl?: string;
}

export interface Ticket {
  id: string;
  userId: number;
  title: string;
  description: string;
  category: TicketCategory;
  status: TicketStatus;
  createdAt: Date;
  updatedAt: Date;
  resolvedAt?: Date;
  attachments?: string[];
}

export enum TicketCategory {
  LOGIN_ERROR = 'Ошибка входа',
  PAYMENT_ISSUE = 'Проблема с оплатой',
  BUG_REPORT = 'Баг в приложении',
  FEATURE_REQUEST = 'Предложение функции',
  OTHER = 'Другое'
}

export enum TicketStatus {
  NEW = 'новая',
  IN_PROGRESS = 'в работе',
  RESOLVED = 'решена',
  CLOSED = 'закрыта'
}

export interface Message {
  id: string;
  ticketId: string;
  senderId: number;
  senderType: 'user' | 'support';
  content: string;
  timestamp: Date;
  attachments?: string[];
}

export interface SupportSpecialist {
  id: number;
  firstName: string;
  lastName?: string;
  username?: string;
  avatar?: string;
}