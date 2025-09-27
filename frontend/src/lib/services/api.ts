import axios, { type AxiosInstance, type AxiosResponse } from 'axios';
import { browser } from '$app/environment';
import type { ApiResponse, ApiError } from '$lib/types/api';

class ApiClient {
	private client: AxiosInstance;

	constructor() {
		// Use environment variables or fallback to development defaults
		const apiUrl = browser
			? (import.meta.env.VITE_API_URL || 'http://localhost:8080')
			: 'http://backend:8080';

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
				if (error.response?.status === 401 && browser) {
					// Token expired, try to refresh
					const refreshToken = localStorage.getItem('refresh_token');
					if (refreshToken) {
						try {
							const response = await this.client.post('/refresh-token', {
								refresh_token: refreshToken
							});

							if (response.data.success && response.data.data) {
								localStorage.setItem('access_token', response.data.data.access_token);
								localStorage.setItem('refresh_token', response.data.data.refresh_token);

								// Retry the original request
								error.config.headers.Authorization = `Bearer ${response.data.data.access_token}`;
								return this.client.request(error.config);
							}
						} catch (refreshError) {
							// Refresh failed, clear tokens and redirect to login
							localStorage.removeItem('access_token');
							localStorage.removeItem('refresh_token');
							window.location.href = '/login';
						}
					} else {
						// No refresh token, redirect to login
						window.location.href = '/login';
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