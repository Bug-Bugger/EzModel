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
			? (import.meta.env.VITE_API_URL || 'http://localhost:8080/api')
			: 'http://backend:8080/api';

		this.client = axios.create({
			baseURL: apiUrl,
			headers: {
				'Content-Type': 'application/json'
			},
			timeout: 10000
		});

		// Request interceptor to add auth token
		this.client.interceptors.request.use(
			(config) => {
				if (browser) {
					const token = localStorage.getItem('access_token');
					if (token) {
						config.headers.Authorization = `Bearer ${token}`;
					}
				}
				return config;
			},
			(error) => Promise.reject(error)
		);

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
						}).then((token) => {
							originalRequest.headers.Authorization = `Bearer ${token}`;
							return this.client(originalRequest);
						}).catch((err) => {
							return Promise.reject(err);
						});
					}

					originalRequest._retry = true;
					this.isRefreshing = true;

					const refreshToken = localStorage.getItem('refresh_token');
					if (refreshToken) {
						try {
							const response = await this.client.post<ApiResponse<LoginResponse>>('/refresh-token', {
								refresh_token: refreshToken
							});

							if (response.data.success && response.data.data) {
								const newTokens = response.data.data;
								localStorage.setItem('access_token', newTokens.access_token);
								localStorage.setItem('refresh_token', newTokens.refresh_token);

								// Process the failed queue
								this.processQueue(null, newTokens.access_token);

								// Retry the original request
								originalRequest.headers.Authorization = `Bearer ${newTokens.access_token}`;
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
					} else {
						// No refresh token, logout immediately
						this.logout();
						return Promise.reject(error);
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

	private logout() {
		localStorage.removeItem('access_token');
		localStorage.removeItem('refresh_token');
		localStorage.removeItem('user');
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