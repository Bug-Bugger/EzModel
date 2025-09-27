import { authStore } from '../stores/auth';
import { browser } from '$app/environment';
import { dev } from '$app/environment';

export interface WebSocketMessage {
	type: string;
	data?: any;
	user_id?: string;
	user_name?: string;
	x?: number;
	y?: number;
	token?: string;
}

export interface WebSocketConfig {
	url: string;
	token: string;
	onMessage: (message: WebSocketMessage) => void;
	onOpen?: () => void;
	onClose?: () => void;
	onError?: (error: Event) => void;
}

export class WebSocketClient {
	private ws: WebSocket | null = null;
	private config: WebSocketConfig;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000;
	private isDestroyed = false;

	constructor(config: WebSocketConfig) {
		this.config = config;
	}

	connect(): Promise<void> {
		return new Promise((resolve, reject) => {
			try {
				this.ws = new WebSocket(this.config.url);

				this.ws.onopen = () => {
					console.log('WebSocket connected');

					// Send authentication immediately
					this.send({
						type: 'authenticate',
						token: this.config.token
					});

					this.reconnectAttempts = 0;
					this.config.onOpen?.();
					resolve();
				};

				this.ws.onmessage = (event) => {
					try {
						const message: WebSocketMessage = JSON.parse(event.data);
						this.config.onMessage(message);
					} catch (error) {
						console.error('Failed to parse WebSocket message:', error);
					}
				};

				this.ws.onclose = (event) => {
					console.log('WebSocket disconnected:', event.code, event.reason);
					this.ws = null;
					this.config.onClose?.();

					// Attempt to reconnect if not destroyed and not a normal closure
					if (!this.isDestroyed && event.code !== 1000) {
						this.scheduleReconnect();
					}
				};

				this.ws.onerror = (error) => {
					console.error('WebSocket error:', error);
					this.config.onError?.(error);
					reject(error);
				};

			} catch (error) {
				console.error('Failed to create WebSocket:', error);
				reject(error);
			}
		});
	}

	private scheduleReconnect() {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			console.log('Max reconnection attempts reached');
			return;
		}

		this.reconnectAttempts++;
		const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

		console.log(`Scheduling reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`);

		setTimeout(() => {
			if (!this.isDestroyed) {
				this.connect().catch(error => {
					console.error('Reconnection failed:', error);
				});
			}
		}, delay);
	}

	send(message: WebSocketMessage): boolean {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			try {
				this.ws.send(JSON.stringify(message));
				return true;
			} catch (error) {
				console.error('Failed to send WebSocket message:', error);
				return false;
			}
		}
		return false;
	}

	disconnect() {
		this.isDestroyed = true;

		if (this.ws) {
			this.ws.close(1000, 'Client disconnecting');
			this.ws = null;
		}
	}

	isConnected(): boolean {
		return this.ws?.readyState === WebSocket.OPEN;
	}

	getReadyState(): number | null {
		return this.ws?.readyState || null;
	}

	// Public methods to update callbacks
	setMessageHandler(handler: (message: WebSocketMessage) => void) {
		this.config.onMessage = handler;
	}

	setCallbacks(callbacks: {
		onOpen?: () => void;
		onClose?: () => void;
		onError?: (error: Event) => void;
	}) {
		if (callbacks.onOpen) this.config.onOpen = callbacks.onOpen;
		if (callbacks.onClose) this.config.onClose = callbacks.onClose;
		if (callbacks.onError) this.config.onError = callbacks.onError;
	}
}

// Factory function to create WebSocket client for collaboration
export function createCollaborationClient(projectId: string): Promise<WebSocketClient> {
	return new Promise((resolve, reject) => {
		// Get auth token from localStorage (same pattern as API client)
		let token: string | null = null;
		if (browser) {
			token = localStorage.getItem('access_token');
		}

		if (!token) {
			reject(new Error('No authentication token available'));
			return;
		}

		// Create WebSocket URL
		let wsUrl: string;

		if (browser) {
			const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			const host = window.location.hostname;
			const port = window.location.port ? `:${window.location.port}` : '';
			wsUrl = `${protocol}//${host}${port}/api/projects/${projectId}/collaborate`;
		} else {
			wsUrl = `ws://backend:8080/api/projects/${projectId}/collaborate`;
		}

		// For development, use localhost backend
		if (dev) {
			wsUrl = `ws://localhost:8080/api/projects/${projectId}/collaborate`;
		}

		const client = new WebSocketClient({
			url: wsUrl,
			token,
			onMessage: () => {}, // Will be set by collaboration store
			onOpen: () => console.log('Collaboration WebSocket connected'),
			onClose: () => console.log('Collaboration WebSocket disconnected'),
			onError: (error) => console.error('Collaboration WebSocket error:', error)
		});

		resolve(client);
	});
}