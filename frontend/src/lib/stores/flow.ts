import { writable } from 'svelte/store';
import type { Node, Edge } from '@xyflow/svelte';
import { projectService } from '$lib/services/project';
import type { Table } from '$lib/types/models';

export interface Position {
	x: number;
	y: number;
}

export interface TableNode extends Node {
	type: 'table';
	data: {
		id: string;
		name: string;
		fields: TableField[];
		position: Position;
	};
}

export interface RelationshipEdge extends Edge {
	type: 'relationship';
	data: {
		id: string;
		fromTable: string;
		toTable: string;
		fromField: string;
		toField: string;
		type: 'one-to-one' | 'one-to-many' | 'many-to-many';
	};
}

export interface TableField {
	id: string;
	name: string;
	type: string;
	is_primary: boolean;
	is_foreign: boolean;
	is_required: boolean;
	is_unique: boolean;
	default_value?: string;
	constraints?: string[];
}

interface FlowState {
	nodes: TableNode[];
	edges: RelationshipEdge[];
	selectedNode: TableNode | null;
	selectedEdge: RelationshipEdge | null;
	viewport: { x: number; y: number; zoom: number };
	isLoading: boolean;
}

function createFlowStore() {
	const initialState: FlowState = {
		nodes: [],
		edges: [],
		selectedNode: null,
		selectedEdge: null,
		viewport: { x: 0, y: 0, zoom: 1 },
		isLoading: false
	};

	const { subscribe, set, update } = writable(initialState);

	return {
		subscribe,

		// Load canvas data from project
		loadCanvasData(canvasData: string) {
			try {
				const data = JSON.parse(canvasData);
				update(state => ({
					...state,
					nodes: data.nodes || [],
					edges: data.edges || [],
					viewport: data.viewport || { x: 0, y: 0, zoom: 1 }
				}));
			} catch (error) {
				console.error('Failed to parse canvas data:', error);
			}
		},

		// Save current canvas state
		getCurrentCanvasData(): string {
			let currentState: FlowState;
			const unsubscribe = subscribe(state => currentState = state);
			unsubscribe();

			return JSON.stringify({
				nodes: currentState!.nodes,
				edges: currentState!.edges,
				viewport: currentState!.viewport
			});
		},

		// Add new table node with API integration
		async addTableNode(
			projectId: string,
			table: Omit<TableNode['data'], 'position' | 'id'>,
			position: Position
		): Promise<TableNode> {
			try {
				// First, persist to backend
				const backendTable: Table = await projectService.createTable(projectId, {
					name: table.name,
					pos_x: position.x,
					pos_y: position.y
				});

				// Create the flow node with backend-generated ID
				const newNode: TableNode = {
					id: backendTable.id,
					type: 'table',
					position,
					data: {
						id: backendTable.id,
						name: backendTable.name,
						fields: table.fields || [],
						position
					}
				};

				// Add to local store
				update(state => ({
					...state,
					nodes: [...state.nodes, newNode]
				}));

				return newNode;
			} catch (error) {
				console.error('Failed to create table:', error);
				throw error;
			}
		},

		// Add table node without API (for loading existing data)
		addLocalTableNode(table: Omit<TableNode['data'], 'position'>, position: Position) {
			const newNode: TableNode = {
				id: table.id,
				type: 'table',
				position,
				data: {
					...table,
					position
				}
			};

			update(state => ({
				...state,
				nodes: [...state.nodes, newNode]
			}));

			return newNode;
		},

		// Update table node locally
		updateTableNode(nodeId: string, updates: Partial<TableNode['data']>) {
			update(state => ({
				...state,
				nodes: state.nodes.map(node =>
					node.id === nodeId
						? { ...node, data: { ...node.data, ...updates } }
						: node
				)
			}));
		},

		// Update table position with API integration
		async updateTablePosition(projectId: string, nodeId: string, position: Position): Promise<void> {
			try {
				// Update backend first
				await projectService.updateTablePosition(projectId, nodeId, {
					pos_x: position.x,
					pos_y: position.y
				});

				// Update local store
				update(state => ({
					...state,
					nodes: state.nodes.map(node =>
						node.id === nodeId
							? {
								...node,
								position,
								data: { ...node.data, position }
							}
							: node
					)
				}));
			} catch (error) {
				console.error('Failed to update table position:', error);
				throw error;
			}
		},

		// Update table position from external source (collaboration)
		updateTablePositionFromExternal(nodeId: string, position: Position) {
			update(state => ({
				...state,
				nodes: state.nodes.map(node =>
					node.id === nodeId
						? {
							...node,
							position,
							data: { ...node.data, position }
						}
						: node
				)
			}));
		},

		// Add table node from external source (collaboration)
		addTableNodeFromExternal(tableData: any, position: Position) {
			const newNode: TableNode = {
				id: tableData.id,
				type: 'table',
				position: position,
				data: {
					id: tableData.id,
					name: tableData.name,
					fields: tableData.fields || [],
					position: position
				}
			};

			update(state => ({
				...state,
				nodes: [...state.nodes, newNode]
			}));

			return newNode;
		},

		// Remove table node with API integration
		async removeTableNode(projectId: string, nodeId: string): Promise<void> {
			try {
				// Delete from backend first
				await projectService.deleteTable(projectId, nodeId);

				// Remove from local store
				update(state => ({
					...state,
					nodes: state.nodes.filter(node => node.id !== nodeId),
					edges: state.edges.filter(edge =>
						edge.source !== nodeId && edge.target !== nodeId
					),
					selectedNode: state.selectedNode?.id === nodeId ? null : state.selectedNode
				}));
			} catch (error) {
				console.error('Failed to delete table:', error);
				throw error;
			}
		},

		// Remove table node locally (for optimistic updates)
		removeLocalTableNode(nodeId: string) {
			update(state => ({
				...state,
				nodes: state.nodes.filter(node => node.id !== nodeId),
				edges: state.edges.filter(edge =>
					edge.source !== nodeId && edge.target !== nodeId
				),
				selectedNode: state.selectedNode?.id === nodeId ? null : state.selectedNode
			}));
		},

		// Add relationship edge
		addRelationshipEdge(relationship: RelationshipEdge['data']) {
			const newEdge: RelationshipEdge = {
				id: relationship.id,
				type: 'relationship',
				source: relationship.fromTable,
				target: relationship.toTable,
				data: relationship
			};

			update(state => ({
				...state,
				edges: [...state.edges, newEdge]
			}));

			return newEdge;
		},

		// Update relationship edge
		updateRelationshipEdge(edgeId: string, updates: Partial<RelationshipEdge['data']>) {
			update(state => ({
				...state,
				edges: state.edges.map(edge =>
					edge.id === edgeId
						? { ...edge, data: { ...edge.data, ...updates } }
						: edge
				)
			}));
		},

		// Remove relationship edge
		removeRelationshipEdge(edgeId: string) {
			update(state => ({
				...state,
				edges: state.edges.filter(edge => edge.id !== edgeId),
				selectedEdge: state.selectedEdge?.id === edgeId ? null : state.selectedEdge
			}));
		},

		// Select node
		selectNode(node: TableNode | null) {
			update(state => ({
				...state,
				selectedNode: node,
				selectedEdge: null
			}));
		},

		// Select edge
		selectEdge(edge: RelationshipEdge | null) {
			update(state => ({
				...state,
				selectedEdge: edge,
				selectedNode: null
			}));
		},

		// Update viewport
		updateViewport(viewport: { x: number; y: number; zoom: number }) {
			update(state => ({
				...state,
				viewport
			}));
		},


		// Clear all data
		clear() {
			set(initialState);
		}
	};
}

export const flowStore = createFlowStore();