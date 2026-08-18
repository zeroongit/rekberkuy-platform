export type EscrowStatus =
  | 'WAITING_PAYMENT'
  | 'FUNDS_LOCKED'
  | 'RELEASED'
  | 'DISPUTED'
  | 'REFUNDED';

export type TransactionType = 'GOODS' | 'SERVICES' | 'EVENTS';

export interface UserProfile {
  id: string;
  email: string;
  fullName: string;
  role: 'USER' | 'VERIFIED_MERCHANT' | 'VERIFIED_VENDOR' | 'ADMIN';
  isKycVerified: boolean;
  avatarUrl?: string;
}

export interface RekberPayWallet {
  userId: string;
  balance: number;
  lockedBalance: number;
  currency: string;
  updatedAt: string;
}

export interface TransactionMilestone {
  id: string;
  title: string;
  amount: number;
  status: 'PENDING' | 'COMPLETED' | 'RELEASED';
  dueDate?: string;
}

export interface Transaction {
  id: string;
  title: string;
  type: TransactionType;
  status: EscrowStatus;
  amount: number;
  buyerId: string;
  sellerId: string;
  buyerName?: string;
  sellerName?: string;
  createdAt: string;
  milestones?: TransactionMilestone[];
  blockchainTxHash?: string;
}

export interface NotificationItem {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
  read: boolean;
  createdAt: string;
}
