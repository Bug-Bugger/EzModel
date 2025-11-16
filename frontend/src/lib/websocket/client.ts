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
	token?: string | null;
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
	private isAuthenticated = false;
	private connectionResolver: ((value: void) => void) | null = null;
	private connectionRejecter: ((reason: Error) => void) | null = null;
	private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;

	constructor(config: WebSocketConfig) {
		this.config = config;
	}

	connect(): Promise<void> {
		return new Promise((resolve, reject) => {
			try {
				// Connect to WebSocket without token in URL (more secure)
				console.log('WebSocket: Connecting to:', this.config.url);
				this.ws = new WebSocket(this.config.url);

				// Store resolve/reject for authentication flow
				this.connectionResolver = resolve;
				this.connectionRejecter = reject;

				this.ws.onopen = () => {
					console.log('WebSocket connected, sending authentication message');

					// Send authentication message after connection is established
					this.send({
						type: 'auth',
						data: this.config.token ? { token: this.config.token } : {}
					});
				};

				this.ws.onmessage = (event) => {
					try {
						// Handle newline-separated JSON messages from backend
						const messages = event.data
							.trim()
							.split('\n')
							.filter((line: string) => line.trim());

						for (const messageText of messages) {
							try {
								const message: WebSocketMessage = JSON.parse(messageText);

								// Handle authentication response
								if (message.type === 'auth' && !this.isAuthenticated) {
									console.log('WebSocket: Authentication successful');
									this.isAuthenticated = true;
									this.reconnectAttempts = 0;

									// Resolve the connection promise
									if (this.connectionResolver) {
										this.connectionResolver();
										this.connectionResolver = null;
										this.connectionRejecter = null;
									}

									// Call onOpen callback after successful authentication
									this.config.onOpen?.();
									continue;
								}

								// Handle authentication errors
								if (message.type === 'error' && !this.isAuthenticated) {
									console.error('WebSocket: Authentication failed:', message.data);
									if (this.connectionRejecter) {
										this.connectionRejecter(new Error('Authentication failed'));
										this.connectionResolver = null;
										this.connectionRejecter = null;
									}
									this.ws?.close(4001, 'Authentication failed');
									continue;
								}

								// Pass other messages to the message handler
								this.config.onMessage(message);
							} catch (parseError) {
								console.error('Failed to parse individual message:', parseError);
								console.error('Message text:', messageText);
							}
						}
					} catch (error) {
						console.error('Failed to handle WebSocket message:', error);
						console.error('Raw data:', event.data);
					}
				};

				this.ws.onclose = (event) => {
					console.log('WebSocket disconnected:', event.code, event.reason);
					this.ws = null;
					this.isAuthenticated = false;

					// Reject connection promise if it's still pending
					if (this.connectionRejecter) {
						this.connectionRejecter(new Error('Connection closed before authentication'));
						this.connectionResolver = null;
						this.connectionRejecter = null;
					}

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

		// Clear any existing reconnect timeout
		if (this.reconnectTimeoutId) {
			clearTimeout(this.reconnectTimeoutId);
			this.reconnectTimeoutId = null;
		}

		this.reconnectAttempts++;
		const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

		console.log(
			`Scheduling reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`
		);

		this.reconnectTimeoutId = setTimeout(() => {
			if (!this.isDestroyed) {
				this.connect().catch((error) => {
					console.error('Reconnection failed:', error);
				});
			}
			this.reconnectTimeoutId = null;
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
		this.isAuthenticated = false;

		// Clear any pending reconnection timeout
		if (this.reconnectTimeoutId) {
			clearTimeout(this.reconnectTimeoutId);
			this.reconnectTimeoutId = null;
		}

		if (this.ws) {
			this.ws.close(1000, 'Client disconnecting');
			this.ws = null;
		}

		// Clean up any pending promises
		if (this.connectionRejecter) {
			this.connectionRejecter(new Error('Disconnected by user'));
			this.connectionResolver = null;
			this.connectionRejecter = null;
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
export function createCollaborationClient(
	projectId: string,
	callbacks?: {
		onMessage?: (message: WebSocketMessage) => void;
		onOpen?: () => void;
		onClose?: () => void;
		onError?: (error: Event) => void;
	}
): Promise<WebSocketClient> {
	return new Promise((resolve, reject) => {
		// Create WebSocket URL
		let wsUrl: string;
		let protocol: string;
		let host: string;

		if (browser) {
			protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
			host = window.location.hostname;
			const port = window.location.port ? `:${window.location.port}` : '';
			wsUrl = `${protocol}//${host}${port}/api/projects/${projectId}/collaborate`;
		} else {
			protocol = 'ws:';
			host = 'backend';
			wsUrl = `ws://backend:8080/api/projects/${projectId}/collaborate`;
		}

		// For development, use dev server which will proxy to backend
		if (dev) {
			wsUrl = `${protocol}//${host}:5173/api/projects/${projectId}/collaborate`;
		}

		console.log('WebSocket: Connecting to:', wsUrl);

		const client = new WebSocketClient({
			url: wsUrl,
			token: null, // Tokens are now sent via secure cookies
			onMessage: callbacks?.onMessage || (() => {}),
			onOpen: callbacks?.onOpen || (() => console.log('Collaboration WebSocket connected')),
			onClose: callbacks?.onClose || (() => console.log('Collaboration WebSocket disconnected')),
			onError:
				callbacks?.onError || ((error) => console.error('Collaboration WebSocket error:', error))
		});

		resolve(client);
	});
}
