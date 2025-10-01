import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { User } from '$lib/types/models';

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

function createAuthStore() {
	const initialState: AuthState = {
		user: null,
		isAuthenticated: false,
		isLoading: true
	};

	const { subscribe, set, update } = writable(initialState);

	return {
		subscribe,

		// Initialize auth state from localStorage
		init() {
			if (!browser) return;

			update((state) => ({ ...state, isLoading: true }));

			const token = localStorage.getItem('access_token');
			const userStr = localStorage.getItem('user');

			if (token && userStr) {
				try {
					const user = JSON.parse(userStr);
					set({
						user,
						isAuthenticated: true,
						isLoading: false
					});
				} catch (error) {
					// Invalid stored user data, clear it
					localStorage.removeItem('access_token');
					localStorage.removeItem('refresh_token');
					localStorage.removeItem('user');
					set({
						user: null,
						isAuthenticated: false,
						isLoading: false
					});
				}
			} else {
				set({
					user: null,
					isAuthenticated: false,
					isLoading: false
				});
			}
		},

		// Set authenticated user
		setUser(user: User) {
			if (browser) {
				localStorage.setItem('user', JSON.stringify(user));
			}
			set({
				user,
				isAuthenticated: true,
				isLoading: false
			});
		},

		// Clear auth state (logout)
		clear() {
			if (browser) {
				localStorage.removeItem('access_token');
				localStorage.removeItem('refresh_token');
				localStorage.removeItem('user');
			}
			set({
				user: null,
				isAuthenticated: false,
				isLoading: false
			});
		},

		// Update loading state
		setLoading(isLoading: boolean) {
			update((state) => ({ ...state, isLoading }));
		}
	};
}

export const authStore = createAuthStore();
