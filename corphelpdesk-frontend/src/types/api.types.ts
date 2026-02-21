export interface ApiResponse<T> {
  status: 'success' | 'error';
  data: T | null;
  error: ApiError | null;
  timestamp: string;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, any>;
}

export interface TicketListFilter {
  status?: TicketStatus;
  priority?: TicketPriority;
  page?: number;
  limit?: number;
}