import { apiClient } from './api';
import type { LoginRequest, RegisterRequest, LoginResponse } from '$lib/types/api';
import type { User } from '$lib/types/models';

export class AuthService {
	async login(credentials: LoginRequest): Promise<User> {
		const response = await apiClient.post<LoginResponse>('/login', credentials);
		if (response.success && response.data) {
			// Store tokens
			localStorage.setItem('access_token', response.data.access_token);
			localStorage.setItem('refresh_token', response.data.refresh_token);

			// Fetch user data using the new token
			const user = await this.fetchCurrentUser();
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

	async refreshToken(refreshToken: string): Promise<LoginResponse> {
		const response = await apiClient.post<LoginResponse>('/refresh-token', {
			refresh_token: refreshToken
		});
		if (response.success && response.data) {
			localStorage.setItem('access_token', response.data.access_token);
			localStorage.setItem('refresh_token', response.data.refresh_token);
			return response.data;
		}
		throw new Error(response.message || 'Token refresh failed');
	}

	logout(): void {
		localStorage.removeItem('access_token');
		localStorage.removeItem('refresh_token');
		localStorage.removeItem('user');
	}

	isAuthenticated(): boolean {
		const token = localStorage.getItem('access_token');
		return !!token;
	}

	getCurrentUser(): User | null {
		const userStr = localStorage.getItem('user');
		return userStr ? JSON.parse(userStr) : null;
	}
}

export const authService = new AuthService();