import { writable } from "svelte/store";
import { authStore } from "./auth";
import {
  createCollaborationClient,
  type WebSocketClient,
} from "../websocket/client";
import { flowStore } from "./flow";

export interface CollaboratorCursor {
  x: number; // Global coordinates
  y: number; // Global coordinates
  timestamp: number;
}

export interface ConnectedUser {
  id: string;
  username: string;
  email: string;
  avatar?: string;
  cursor?: CollaboratorCursor;
  lastActivity: number;
}

export interface ActivityEvent {
  id: string;
  userId: string;
  userName: string;
  type:
    | "user_joined"
    | "table_create"
    | "table_update"
    | "table_delete"
    | "field_create"
    | "field_update"
    | "field_delete"
    | "relationship_create"
    | "relationship_delete";
  message: string;
  timestamp: number;
  data?: any;
}

interface CollaborationState {
  isConnected: boolean;
  connectedUsers: ConnectedUser[];
  activityEvents: ActivityEvent[];
  connectionStatus: "connecting" | "connected" | "disconnected" | "error";
  lastError?: string;
  currentUserCursor?: CollaboratorCursor; // Track current user's cursor position locally
}

function createCollaborationStore() {
  const initialState: CollaborationState = {
    isConnected: false,
    connectedUsers: [],
    activityEvents: [],
    connectionStatus: "disconnected",
  };

  const { subscribe, set, update } = writable(initialState);

  let wsClient: WebSocketClient | null = null;

  return {
    subscribe,

    // Connect to WebSocket for collaboration
    async connect(projectId: string) {
      update((state) => ({ ...state, connectionStatus: "connecting" }));

      try {
        // Create WebSocket client with callbacks set up front
        wsClient = await createCollaborationClient(projectId, {
          onMessage: handleWebSocketMessage,
          onOpen: () => {
            console.log("Collaboration WebSocket connected - updating store");
            update((state) => ({
              ...state,
              isConnected: true,
              connectionStatus: "connected",
              lastError: undefined,
            }));
          },
          onClose: () => {
            console.log(
              "Collaboration WebSocket disconnected - updating store"
            );
            update((state) => ({
              ...state,
              isConnected: false,
              connectionStatus: "disconnected",
              connectedUsers: [],
            }));
          },
          onError: (error) => {
            console.error(
              "Collaboration WebSocket error - updating store:",
              error
            );
            update((state) => ({
              ...state,
              connectionStatus: "error",
              lastError: "Connection error occurred",
            }));
          },
        });

        // Initiate connection
        await wsClient.connect();
      } catch (error) {
        console.error("Failed to connect WebSocket:", error);
        update((state) => ({
          ...state,
          connectionStatus: "error",
          lastError: "Failed to establish connection",
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
      // Update local current user cursor position
      update((state) => ({
        ...state,
        currentUserCursor: {
          x,
          y,
          timestamp: Date.now(),
        },
      }));

      // Send to other users via WebSocket
      if (wsClient && wsClient.isConnected()) {
        const message = {
          type: "user_cursor",
          data: {
            cursor_x: x,
            cursor_y: y,
          },
        };
        wsClient.send(message);
      }
    },

    // Send schema change event
    sendSchemaEvent(type: string, data: any) {
      if (wsClient && wsClient.isConnected()) {
        wsClient.send({
          type,
          data,
        });
      }
    },

    // Clear activity events
    clearActivity() {
      update((state) => ({ ...state, activityEvents: [] }));
    },
  };

  // Helper function to get username from user_id
  function getUsernameFromId(userId: string, state: CollaborationState): string {
    const user = state.connectedUsers.find(u => u.id === userId);
    return user?.username || "Unknown User";
  }

  function handleWebSocketMessage(message: any) {
    switch (message.type) {
      case "user_joined":
        // Handle backend UserJoinedPayload structure
        const joinedUser = {
          id: message.data.user_id,
          username: message.data.username || "Unknown User",
          email: "", // Not provided in the payload
          lastActivity: Date.now(),
        };

        update((state) => ({
          ...state,
          connectedUsers: [...state.connectedUsers, joinedUser],
        }));

        addActivityEvent({
          type: "user_joined",
          userId: message.data.user_id,
          userName: message.data.username || "Unknown User",
          message: `${
            message.data.username || "Unknown User"
          } joined the collaboration`,
        });
        break;

      case "user_left":
        // Handle backend UserLeftPayload structure
        update((state) => ({
          ...state,
          connectedUsers: state.connectedUsers.filter(
            (u) => u.id !== message.data.user_id
          ),
        }));
        break;

      case "user_presence":
        // Handle backend UserPresencePayload structure - this sets the complete user list
        const activeUsers =
          message.data.active_users?.map((user: any) => ({
            id: user.user_id,
            username: user.username || "Unknown User",
            email: "", // Not provided in the payload
            lastActivity: Date.now(),
          })) || [];

        update((state) => ({
          ...state,
          connectedUsers: activeUsers,
        }));
        break;

      case "user_cursor":
        // Handle backend UserCursorPayload structure
        update((state) => {
          const updatedUsers = state.connectedUsers.map((user) =>
            user.id === message.data.user_id
              ? {
                  ...user,
                  cursor: {
                    x: message.data.cursor_x,
                    y: message.data.cursor_y,
                    timestamp: Date.now(),
                  },
                  lastActivity: Date.now(),
                }
              : user
          );

          return {
            ...state,
            connectedUsers: updatedUsers,
          };
        });
        break;

      case "table_create":
      case "table_update":
      case "table_delete":
      case "field_create":
      case "field_update":
      case "field_delete":
      case "relationship_create":
      case "relationship_delete":
        update((state) => {
          const userName = getUsernameFromId(message.user_id, state);
          const newEvent: ActivityEvent = {
            id: crypto.randomUUID(),
            type: message.type,
            userId: message.user_id,
            userName: userName,
            message: generateActivityMessage(message.type, message.data),
            data: message.data,
            timestamp: Date.now(),
          };

          return {
            ...state,
            activityEvents: [newEvent, ...state.activityEvents.slice(0, 49)], // Keep last 50 events
          };
        });
        break;

      case "table_updated":
        // Handle table position updates from other users
        if (message.data.table_id && message.data.x !== undefined && message.data.y !== undefined) {
          // Update table position using static import
          flowStore.updateTablePositionFromExternal(message.data.table_id, {
            x: message.data.x,
            y: message.data.y
          });

          // No activity event for real-time position updates during dragging
          // Activity events will be created only on drag completion
        }
        break;

      case "table_moved":
        // Handle final table position after drag completion
        if (message.data.table_id && message.data.x !== undefined && message.data.y !== undefined) {
          // Update table position using static import
          flowStore.updateTablePositionFromExternal(message.data.table_id, {
            x: message.data.x,
            y: message.data.y
          });

          // Add activity event for completed table move
          update((state) => {
            const userName = getUsernameFromId(message.user_id, state);
            const newEvent: ActivityEvent = {
              id: crypto.randomUUID(),
              type: "table_update", // Use existing activity type for consistency
              userId: message.user_id,
              userName: userName,
              message: `moved table "${message.data.name || 'Unknown Table'}"`,
              data: message.data,
              timestamp: Date.now(),
            };

            return {
              ...state,
              activityEvents: [newEvent, ...state.activityEvents.slice(0, 49)], // Keep last 50 events
            };
          });
        }
        break;

      default:
        console.warn("Unknown message type:", message.type);
    }
  }

  function addActivityEvent(event: Omit<ActivityEvent, "id" | "timestamp">) {
    const newEvent: ActivityEvent = {
      ...event,
      id: crypto.randomUUID(),
      timestamp: Date.now(),
    };

    update((state) => ({
      ...state,
      activityEvents: [newEvent, ...state.activityEvents.slice(0, 49)], // Keep last 50 events
    }));
  }

  function generateActivityMessage(type: string, data: any): string {
    switch (type) {
      case "table_create":
        return `created table "${data.name}"`;
      case "table_update":
        return `updated table "${data.name}"`;
      case "table_delete":
        return `deleted table "${data.name}"`;
      case "field_create":
        return `added field "${data.name}" to "${data.table_name}"`;
      case "field_update":
        return `updated field "${data.name}" in "${data.table_name}"`;
      case "field_delete":
        return `removed field "${data.name}" from "${data.table_name}"`;
      case "relationship_create":
        return `created relationship between "${data.from_table}" and "${data.to_table}"`;
      case "relationship_delete":
        return `removed relationship between "${data.from_table}" and "${data.to_table}"`;
      default:
        return "made a change";
    }
  }
}

export const collaborationStore = createCollaborationStore();
