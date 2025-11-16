import { apiClient } from './api';
import type { LoginRequest, RegisterRequest, LoginResponse } from '$lib/types/api';
import type { User } from '$lib/types/models';

export class AuthService {
	async login(credentials: LoginRequest): Promise<User> {
		// Tokens are now set as httpOnly cookies by the backend
		const response = await apiClient.post<{ user: User }>('/login', credentials);
		if (response.success && response.data) {
			// Backend now returns user data directly
			const user = response.data.user;
			localStorage.setItem('user', JSON.stringify(user));
			return user;
		}
		throw new Error(response.message || 'Login failed');
	}

	async fetchCurrentUser(): Promise<User> {
		const response = await apiClient.get<User>('/me');
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Failed to get current user');
	}

	async register(userData: RegisterRequest): Promise<User> {
		const response = await apiClient.post<User>('/register', userData);
		if (response.success && response.data) {
			return response.data;
		}
		throw new Error(response.message || 'Registration failed');
	}

	async logout(): Promise<void> {
		// Call backend to clear httpOnly cookies
		try {
			await apiClient.post('/logout');
		} catch (error) {
			console.error('Logout error:', error);
		}
		// Clear local user data
		localStorage.removeItem('user');
	}

	isAuthenticated(): boolean {
		// Check if user data exists in localStorage
		// Actual authentication is handled by httpOnly cookies
		const userStr = localStorage.getItem('user');
		return !!userStr;
	}

	getCurrentUser(): User | null {
		const userStr = localStorage.getItem('user');
		return userStr ? JSON.parse(userStr) : null;
	}
}

export const authService = new AuthService();
