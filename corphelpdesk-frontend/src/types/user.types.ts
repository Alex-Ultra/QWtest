export type VerificationStatus = 'pending' | 'verified' | 'rejected';
export type UserRole = 'client' | 'agent' | 'admin' | 'security_officer';

export interface User {
  id: string;
  telegramId: string;
  fullName: string;
  username?: string;
  role: UserRole;
  verificationStatus: VerificationStatus;
  rejectionReason?: string;
  pdConsentAccepted: boolean;
  createdAt: string;
}

export interface UserProfileUpdate {
  snils: string;
  dateOfBirth: string; // Format: YYYY-MM-DD
  phone: string;
  pdConsent: boolean;
}