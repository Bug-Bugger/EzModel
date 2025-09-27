import { apiClient } from './api';
import type { User } from '$lib/types/models';

export class UserService {
	async getAllUsers(): Promise<User[]> {
		const response = await apiClient.get<User[]>('/users');
		if (response.success && response.data) {
			return response.data;
		}
		return [];
	}

	async searchUsers(query: string): Promise<User[]> {
		const users = await this.getAllUsers();
		return users.filter(user =>
			(user.email || '').toLowerCase().includes(query.toLowerCase()) ||
			(user.username || '').toLowerCase().includes(query.toLowerCase())
		);
	}
}

export const userService = new UserService();