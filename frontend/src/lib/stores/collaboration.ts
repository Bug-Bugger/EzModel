import { writable, get } from "svelte/store";
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
    | "table_created"
    | "table_update"
    | "table_deleted"
    | "field_created"
    | "field_updated"
    | "field_deleted"
    | "relationship_create"
    | "relationship_update"
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
      console.log("📤 Sending schema event:", type, data);
      if (wsClient && wsClient.isConnected()) {
        const message = {
          type,
          data,
        };
        console.log("📡 WebSocket sending:", message);
        wsClient.send(message);
      } else {
        console.warn("⚠️ WebSocket not connected, cannot send schema event:", type, data);
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
    console.log("🔄 WebSocket message received:", message.type, message);

    // Check if this message is from the current user to prevent duplicate updates
    const currentUser = get(authStore).user;
    const isOwnMessage = currentUser && message.user_id === currentUser.id;

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

      case "table_created":
        // Handle table creation - add table to canvas and create activity
        if (message.data.id && message.data.name && message.data.pos_x !== undefined && message.data.pos_y !== undefined) {
          // Add table to canvas using the flow store
          flowStore.addTableNodeFromExternal(message.data, {
            x: message.data.pos_x,
            y: message.data.pos_y
          });
        }

        // Create activity event for table creation
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

      case "table_deleted":
        console.log("🗑️ Received table_deleted message:", message);
        console.log("📊 Message data:", message.data);
        console.log("👤 Message user_id:", message.user_id);

        try {
          // Handle table deletion - remove table from canvas
          if (message.data && message.data.id) {
            console.log("🔄 Removing table with ID:", message.data.id);
            flowStore.removeLocalTableNode(message.data.id);
            console.log("✅ Table removal completed");
          } else {
            console.warn("⚠️ Missing table ID in deletion message:", message.data);
          }

          // Create activity event for table deletion
          update((state) => {
            try {
              const userName = message.user_id ? getUsernameFromId(message.user_id, state) : "Unknown User";
              console.log("👤 Resolved username:", userName);

              const newEvent: ActivityEvent = {
                id: crypto.randomUUID(),
                type: message.type,
                userId: message.user_id || "unknown",
                userName: userName,
                message: generateActivityMessage(message.type, message.data || {}),
                data: message.data || {},
                timestamp: Date.now(),
              };

              console.log("📝 Creating activity event:", newEvent);

              return {
                ...state,
                activityEvents: [newEvent, ...state.activityEvents.slice(0, 49)], // Keep last 50 events
              };
            } catch (activityError) {
              console.error("❌ Error creating activity event:", activityError);
              return state; // Return unchanged state on error
            }
          });
        } catch (error) {
          console.error("❌ Error handling table_deleted message:", error, message);
        }
        break;

      case "table_update":
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

      case "relationship_create":
        // Add relationship edge to flow store for real-time collaboration
        if (message.data.id && message.data.source_table_id && message.data.target_table_id) {
          flowStore.addLocalRelationshipEdge({
            id: message.data.id,
            fromTable: message.data.source_table_id,
            toTable: message.data.target_table_id,
            fromField: message.data.source_field_id,
            toField: message.data.target_field_id,
            type: message.data.relation_type
          });
        }

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

      case "relationship_update":
        // Update relationship edge in flow store for real-time collaboration
        if (message.data.id) {
          flowStore.updateLocalRelationshipEdge(message.data.id, {
            type: message.data.relation_type
          });
        }

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

      case "relationship_delete":
        // Remove relationship edge from flow store for real-time collaboration
        if (message.data.id) {
          flowStore.removeLocalRelationshipEdge(message.data.id);
        }

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

      case "field_created":
        // Handle field creation - add field to table and create activity
        if (message.data.table_id && message.data.field_id) {
          // Only update local state if this is NOT from the current user (prevent duplicates)
          if (!isOwnMessage) {
            // Use backend field data structure directly
            const fieldData = {
              id: message.data.field_id,
              table_id: message.data.table_id,
              name: message.data.name,
              data_type: message.data.type,
              is_primary_key: message.data.is_primary || false,
              is_nullable: message.data.is_nullable,
              default_value: message.data.default || "",
              position: message.data.position || 0,
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            };

            // Update the table's fields in the flow store
            const currentNodes = get(flowStore).nodes;
            const tableNode = currentNodes.find(node => node.id === message.data.table_id);
            if (tableNode) {
              const updatedFields = [...tableNode.data.fields, fieldData];
              flowStore.updateTableNode(tableNode.id, { fields: updatedFields });
            }
          }
        }

        // Create activity event
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

      case "field_updated":
        // Handle field update - update field in table and create activity
        if (message.data.table_id && message.data.field_id) {
          // Only update local state if this is NOT from the current user (prevent duplicates)
          if (!isOwnMessage) {
            // Use backend field data structure directly
            const fieldUpdates = {
              name: message.data.name,
              data_type: message.data.type,
              is_primary_key: message.data.is_primary || false,
              is_nullable: message.data.is_nullable,
              default_value: message.data.default || "",
              position: message.data.position || 0,
              updated_at: new Date().toISOString()
            };

            // Update the field in the flow store
            const currentNodes = get(flowStore).nodes;
            const tableNode = currentNodes.find(node => node.id === message.data.table_id);
            if (tableNode) {
              const updatedFields = tableNode.data.fields.map(field =>
                field.id === message.data.field_id ? { ...field, ...fieldUpdates } : field
              );
              flowStore.updateTableNode(tableNode.id, { fields: updatedFields });
            }
          }
        }

        // Create activity event
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

      case "field_deleted":
        // Handle field deletion - remove field from table and create activity
        if (message.data.table_id && message.data.field_id) {
          // Only update local state if this is NOT from the current user (prevent duplicates)
          if (!isOwnMessage) {
            // Remove the field from the flow store
            const currentNodes = get(flowStore).nodes;
            const tableNode = currentNodes.find(node => node.id === message.data.table_id);
            if (tableNode) {
              const updatedFields = tableNode.data.fields.filter(field => field.id !== message.data.field_id);
              flowStore.updateTableNode(tableNode.id, { fields: updatedFields });
            }
          }
        }

        // Create activity event
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
        // Handle table position updates during drag (visual only, no activity entries)
        if (message.data.table_id && message.data.x !== undefined && message.data.y !== undefined) {
          // Update table position using static import
          flowStore.updateTablePositionFromExternal(message.data.table_id, {
            x: message.data.x,
            y: message.data.y
          });
          // Note: No activity event created for table moves to avoid spamming activity feed
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
      case "table_created":
        return `created table "${data.name}"`;
      case "table_update":
        return `updated table "${data.name}"`;
      case "table_deleted":
        return `deleted table "${data.name}"`;
      case "field_created":
        return `added field "${data.name}" to table`;
      case "field_updated":
        return `updated field "${data.name}" in table`;
      case "field_deleted":
        return `removed field "${data.name}" from table`;
      case "relationship_create":
        return `created relationship between "${data.from_table}" and "${data.to_table}"`;
      case "relationship_update":
        return `updated relationship between "${data.from_table}" and "${data.to_table}"`;
      case "relationship_delete":
        return `removed relationship between "${data.from_table}" and "${data.to_table}"`;
      default:
        return "made a change";
    }
  }
}

export const collaborationStore = createCollaborationStore();
