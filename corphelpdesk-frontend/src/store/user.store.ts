import { create } from 'zustand';
import { User } from '../types/user.types';

interface UserStore {
  user: User | null;
  setUser: (user: User) => void;
  clearUser: () => void;
  updateProfile: (profile: Partial<User>) => void;
}

export const useUserStore = create<UserStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  clearUser: () => set({ user: null }),
  updateProfile: (profile) => set((state) => ({
    user: state.user ? { ...state.user, ...profile } : null
  }))
}));