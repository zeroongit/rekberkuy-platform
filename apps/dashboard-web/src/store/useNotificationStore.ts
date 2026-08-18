import { create } from 'zustand';
import { NotificationItem } from '@/types';

interface NotificationState {
  notifications: NotificationItem[];
  addNotification: (notification: Omit<NotificationItem, 'id' | 'createdAt' | 'read'>) => void;
  markAsRead: (id: string) => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  notifications: [
    {
      id: 'notif-1',
      title: 'Escrow Dana Dikunci',
      message: 'Dana sebesar Rp 500.000 untuk transaksi #TRX-9821 telah berhasil dikunci.',
      type: 'success',
      read: false,
      createdAt: new Date().toISOString(),
    },
  ],
  addNotification: (notif) =>
    set((state) => ({
      notifications: [
        {
          ...notif,
          id: 'notif-' + Date.now(),
          read: false,
          createdAt: new Date().toISOString(),
        },
        ...state.notifications,
      ],
    })),
  markAsRead: (id) =>
    set((state) => ({
      notifications: state.notifications.map((n) => (n.id === id ? { ...n, read: true } : n)),
    })),
  clearAll: () => set({ notifications: [] }),
}));
