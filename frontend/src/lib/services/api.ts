import axios, { type AxiosInstance, type AxiosResponse } from 'axios';
import { browser } from '$app/environment';
import type { ApiResponse, ApiError, LoginResponse } from '$lib/types/api';
import { authStore } from '$lib/stores/auth';

class ApiClient {
	private client: AxiosInstance;
	private isRefreshing = false;
	private failedQueue: Array<{
		resolve: (value: any) => void;
		reject: (error: any) => void;
	}> = [];

	constructor() {
		// Use environment variables or fallback to development defaults
		const apiUrl = browser
			? import.meta.env.VITE_API_URL || 'http://localhost:8080/api'
			: 'http://backend:8080/api';

		this.client = axios.create({
			baseURL: apiUrl,
			headers: {
				'Content-Type': 'application/json'
			},
			timeout: 10000,
			withCredentials: true // Include cookies in requests
		});

		// Response interceptor for error handling
		this.client.interceptors.response.use(
			(response: AxiosResponse<ApiResponse>) => response,
			async (error) => {
				const originalRequest = error.config;

				if (error.response?.status === 401 && browser && !originalRequest._retry) {
					if (this.isRefreshing) {
						// Another request is already refreshing tokens, queue this request
						return new Promise((resolve, reject) => {
							this.failedQueue.push({ resolve, reject });
						})
							.then(() => {
								return this.client(originalRequest);
							})
							.catch((err) => {
								return Promise.reject(err);
							});
					}

					originalRequest._retry = true;
					this.isRefreshing = true;

					try {
						// Call refresh endpoint (cookies are sent automatically)
						const response = await this.client.post<ApiResponse<LoginResponse>>('/refresh-token');

						if (response.data.success) {
							// Tokens are now in httpOnly cookies, no need to store them
							// Process the failed queue
							// No error, no token needed for cookie-based auth
							this.processQueue(/* error */ null, /* token */ null);

							// Retry the original request
							return this.client(originalRequest);
						}
					} catch (refreshError) {
						// Refresh failed, clear tokens and logout
						this.processQueue(refreshError, null);
						this.logout();
						return Promise.reject(refreshError);
					} finally {
						this.isRefreshing = false;
					}
				}

				const apiError: ApiError = {
					message: error.response?.data?.message || error.message || 'An error occurred',
					status: error.response?.status || 0,
					code: error.response?.data?.code
				};

				return Promise.reject(apiError);
			}
		);
	}

	private processQueue(error: any, token: string | null) {
		this.failedQueue.forEach(({ resolve, reject }) => {
			if (error) {
				reject(error);
			} else {
				resolve(token);
			}
		});

		this.failedQueue = [];
	}

	private async logout() {
		try {
			// Call logout endpoint to clear httpOnly cookies on server
			await this.client.post('/logout');
		} catch (error) {
			// Ignore errors during logout
			console.error('Logout error:', error);
		}

		// Clear local user data
		authStore.clear();
		if (browser) {
			window.location.href = '/login';
		}
	}

	async get<T>(url: string): Promise<ApiResponse<T>> {
		const response = await this.client.get<ApiResponse<T>>(url);
		return response.data;
	}

	async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
		const response = await this.client.post<ApiResponse<T>>(url, data);
		return response.data;
	}

	async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
		const response = await this.client.put<ApiResponse<T>>(url, data);
		return response.data;
	}

	async delete<T>(url: string): Promise<ApiResponse<T>> {
		const response = await this.client.delete<ApiResponse<T>>(url);
		return response.data;
	}
}

export const apiClient = new ApiClient();
