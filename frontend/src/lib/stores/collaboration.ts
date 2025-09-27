import { writable } from 'svelte/store';
import { authStore } from './auth';
import { createCollaborationClient, type WebSocketClient } from '../websocket/client';

export interface CollaboratorCursor {
	x: number;
	y: number;
	timestamp: number;
}

export interface ConnectedUser {
	id: string;
	name: string;
	email: string;
	avatar?: string;
	cursor?: CollaboratorCursor;
	lastActivity: number;
}

export interface ActivityEvent {
	id: string;
	userId: string;
	userName: string;
	type: 'user_joined' | 'table_create' | 'table_update' | 'table_delete' | 'field_create' | 'field_update' | 'field_delete' | 'relationship_create' | 'relationship_delete';
	message: string;
	timestamp: number;
	data?: any;
}

interface CollaborationState {
	isConnected: boolean;
	connectedUsers: ConnectedUser[];
	activityEvents: ActivityEvent[];
	connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error';
	lastError?: string;
}

function createCollaborationStore() {
	const initialState: CollaborationState = {
		isConnected: false,
		connectedUsers: [],
		activityEvents: [],
		connectionStatus: 'disconnected'
	};

	const { subscribe, set, update } = writable(initialState);

	let wsClient: WebSocketClient | null = null;

	return {
		subscribe,

		// Connect to WebSocket for collaboration
		async connect(projectId: string) {
			update(state => ({ ...state, connectionStatus: 'connecting' }));

			try {
				// Create WebSocket client with callbacks set up front
				wsClient = await createCollaborationClient(projectId, {
					onMessage: handleWebSocketMessage,
					onOpen: () => {
						console.log('Collaboration WebSocket connected - updating store');
						update(state => ({
							...state,
							isConnected: true,
							connectionStatus: 'connected',
							lastError: undefined
						}));
					},
					onClose: () => {
						console.log('Collaboration WebSocket disconnected - updating store');
						update(state => ({
							...state,
							isConnected: false,
							connectionStatus: 'disconnected',
							connectedUsers: []
						}));
					},
					onError: (error) => {
						console.error('Collaboration WebSocket error - updating store:', error);
						update(state => ({
							...state,
							connectionStatus: 'error',
							lastError: 'Connection error occurred'
						}));
					}
				});

				// Initiate connection
				await wsClient.connect();

			} catch (error) {
				console.error('Failed to connect WebSocket:', error);
				update(state => ({
					...state,
					connectionStatus: 'error',
					lastError: 'Failed to establish connection'
				}));
				throw error;
			}
		},

		// Disconnect WebSocket
		disconnect() {
			if (wsClient) {
				wsClient.disconnect();
				wsClient = null;
			}

			set(initialState);
		},

		// Send cursor position
		sendCursorPosition(x: number, y: number) {
			if (wsClient && wsClient.isConnected()) {
				wsClient.send({
					type: 'cursor_move',
					x,
					y
				});
			}
		},

		// Send schema change event
		sendSchemaEvent(type: string, data: any) {
			if (wsClient && wsClient.isConnected()) {
				wsClient.send({
					type,
					data
				});
			}
		},

		// Clear activity events
		clearActivity() {
			update(state => ({ ...state, activityEvents: [] }));
		}
	};

	function handleWebSocketMessage(message: any) {
		switch (message.type) {
			case 'user_joined':
				update(state => ({
					...state,
					connectedUsers: [...state.connectedUsers, {
						...message.user,
						lastActivity: Date.now()
					}]
				}));
				addActivityEvent({
					type: 'user_joined',
					userId: message.user.id,
					userName: message.user.name,
					message: `${message.user.name} joined the collaboration`
				});
				break;

			case 'user_left':
				update(state => ({
					...state,
					connectedUsers: state.connectedUsers.filter(u => u.id !== message.user_id)
				}));
				break;

			case 'cursor_move':
				update(state => ({
					...state,
					connectedUsers: state.connectedUsers.map(user =>
						user.id === message.user_id
							? {
								...user,
								cursor: {
									x: message.x,
									y: message.y,
									timestamp: Date.now()
								},
								lastActivity: Date.now()
							}
							: user
					)
				}));
				break;

			case 'table_create':
			case 'table_update':
			case 'table_delete':
			case 'field_create':
			case 'field_update':
			case 'field_delete':
			case 'relationship_create':
			case 'relationship_delete':
				addActivityEvent({
					type: message.type,
					userId: message.user_id,
					userName: message.user_name,
					message: generateActivityMessage(message.type, message.data),
					data: message.data
				});
				break;

			default:
				console.log('Unknown message type:', message.type);
		}
	}

	function addActivityEvent(event: Omit<ActivityEvent, 'id' | 'timestamp'>) {
		const newEvent: ActivityEvent = {
			...event,
			id: crypto.randomUUID(),
			timestamp: Date.now()
		};

		update(state => ({
			...state,
			activityEvents: [newEvent, ...state.activityEvents.slice(0, 49)] // Keep last 50 events
		}));
	}

	function generateActivityMessage(type: string, data: any): string {
		switch (type) {
			case 'table_create':
				return `created table "${data.name}"`;
			case 'table_update':
				return `updated table "${data.name}"`;
			case 'table_delete':
				return `deleted table "${data.name}"`;
			case 'field_create':
				return `added field "${data.name}" to "${data.table_name}"`;
			case 'field_update':
				return `updated field "${data.name}" in "${data.table_name}"`;
			case 'field_delete':
				return `removed field "${data.name}" from "${data.table_name}"`;
			case 'relationship_create':
				return `created relationship between "${data.from_table}" and "${data.to_table}"`;
			case 'relationship_delete':
				return `removed relationship between "${data.from_table}" and "${data.to_table}"`;
			default:
				return 'made a change';
		}
	}
}

export const collaborationStore = createCollaborationStore();