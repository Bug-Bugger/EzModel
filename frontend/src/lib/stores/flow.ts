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
		table_id: string;
		name: string;
		fields: TableField[];
		position: Position;
	};
}

export interface RelationshipEdge extends Edge {
	type: 'relationship' | 'default'; // Allow default type for testing
	sourceHandle?: string;
	targetHandle?: string;
	data: {
		relationship_id: string;
		source_table_id: string;
		target_table_id: string;
		source_field_id: string;
		target_field_id: string;
		relation_type: 'one_to_one' | 'one_to_many' | 'many_to_many';
	};
}

export interface TableField {
	field_id: string;
	table_id: string;
	name: string;
	data_type: string;
	is_primary_key: boolean;
	is_nullable: boolean;
	default_value: string;
	position: number;
	created_at: string;
	updated_at: string;
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
				update((state) => ({
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
			const unsubscribe = subscribe((state) => (currentState = state));
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
			table: Omit<TableNode['data'], 'position' | 'table_id'>,
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
					id: backendTable.table_id,
					type: 'table',
					position,
					data: {
						table_id: backendTable.table_id,
						name: backendTable.name,
						fields: table.fields || [],
						position
					}
				};

				// Add to local store
				update((state) => ({
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
				id: table.table_id,
				type: 'table',
				position,
				data: {
					...table,
					position
				}
			};

			update((state) => ({
				...state,
				nodes: [...state.nodes, newNode]
			}));

			return newNode;
		},

		// Update table node locally
		updateTableNode(nodeId: string, updates: Partial<TableNode['data']>) {
			update((state) => {
				const updatedNodes = state.nodes.map((node) =>
					node.id === nodeId ? { ...node, data: { ...node.data, ...updates } } : node
				);

				// Update selectedNode if it's the updated node to ensure proper reactivity
				const updatedSelectedNode =
					state.selectedNode?.id === nodeId
						? updatedNodes.find((n) => n.id === nodeId) || null
						: state.selectedNode;

				return {
					...state,
					nodes: updatedNodes,
					selectedNode: updatedSelectedNode
				};
			});
		},

		// Update table position with API integration
		async updateTablePosition(
			projectId: string,
			nodeId: string,
			position: Position
		): Promise<void> {
			try {
				// Update backend first
				await projectService.updateTablePosition(projectId, nodeId, {
					pos_x: position.x,
					pos_y: position.y
				});

				// Update local store
				update((state) => ({
					...state,
					nodes: state.nodes.map((node) =>
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
			update((state) => ({
				...state,
				nodes: state.nodes.map((node) =>
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
				id: tableData.table_id,
				type: 'table',
				position: position,
				data: {
					table_id: tableData.table_id,
					name: tableData.name,
					fields: tableData.fields || [],
					position: position
				}
			};

			update((state) => ({
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
				update((state) => ({
					...state,
					nodes: state.nodes.filter((node) => node.id !== nodeId),
					edges: state.edges.filter((edge) => edge.source !== nodeId && edge.target !== nodeId),
					selectedNode: state.selectedNode?.id === nodeId ? null : state.selectedNode
				}));
			} catch (error) {
				console.error('Failed to delete table:', error);
				throw error;
			}
		},

		// Remove table node locally (for optimistic updates)
		removeLocalTableNode(nodeId: string) {
			update((state) => ({
				...state,
				nodes: state.nodes.filter((node) => node.id !== nodeId),
				edges: state.edges.filter((edge) => edge.source !== nodeId && edge.target !== nodeId),
				selectedNode: state.selectedNode?.id === nodeId ? null : state.selectedNode
			}));
		},

		// Add relationship edge with API integration
		async addRelationshipEdge(
			projectId: string,
			relationshipData: {
				source_table_id: string;
				source_field_id: string;
				target_table_id: string;
				target_field_id: string;
				relation_type: 'one_to_one' | 'one_to_many' | 'many_to_many';
			}
		): Promise<RelationshipEdge> {
			try {
				// Create via API first
				const newRelationship = await projectService.createRelationship(
					projectId,
					relationshipData
				);

				// Create the flow edge with backend-generated ID
				const edgeData = {
					relationship_id: newRelationship.relationship_id,
					source_table_id: newRelationship.source_table_id,
					target_table_id: newRelationship.target_table_id,
					source_field_id: newRelationship.source_field_id,
					target_field_id: newRelationship.target_field_id,
					relation_type: newRelationship.relation_type
				};

				// Add to local store
				const newEdge = this.addLocalRelationshipEdge(edgeData);
				return newEdge;
			} catch (error) {
				console.error('Failed to create relationship:', error);
				throw error;
			}
		},

		// Add relationship edge without API (for loading existing data)
		addLocalRelationshipEdge(relationship: RelationshipEdge['data']) {
			const newEdge: RelationshipEdge = {
				id: relationship.relationship_id,
				type: 'relationship',
				source: relationship.source_table_id,
				target: relationship.target_table_id,
				sourceHandle: `${relationship.source_table_id}-${relationship.source_field_id}-source`,
				targetHandle: `${relationship.target_table_id}-${relationship.target_field_id}-target`,
				data: relationship
			};

			update((state) => ({
				...state,
				edges: [...state.edges, newEdge]
			}));

			return newEdge;
		},

		// Update relationship edge with API integration
		async updateRelationshipEdge(
			projectId: string,
			edgeId: string,
			updates: {
				relation_type?: 'one_to_one' | 'one_to_many' | 'many_to_many';
			}
		): Promise<void> {
			try {
				// Update via API first
				const updatedRelationship = await projectService.updateRelationship(
					projectId,
					edgeId,
					updates
				);

				// Update local store
				update((state) => ({
					...state,
					edges: state.edges.map((edge) =>
						edge.id === edgeId
							? {
									...edge,
									data: {
										...edge.data,
										relation_type: updatedRelationship.relation_type
									}
								}
							: edge
					),
					selectedEdge:
						state.selectedEdge?.id === edgeId
							? {
									...state.selectedEdge,
									data: {
										...state.selectedEdge.data,
										relation_type: updatedRelationship.relation_type
									}
								}
							: state.selectedEdge
				}));
			} catch (error) {
				console.error('Failed to update relationship:', error);
				throw error;
			}
		},

		// Update relationship edge locally
		updateLocalRelationshipEdge(edgeId: string, updates: Partial<RelationshipEdge['data']>) {
			update((state) => ({
				...state,
				edges: state.edges.map((edge) =>
					edge.id === edgeId ? { ...edge, data: { ...edge.data, ...updates } } : edge
				)
			}));
		},

		// Remove relationship edge with API integration
		async removeRelationshipEdge(projectId: string, edgeId: string): Promise<void> {
			try {
				// Delete from backend first
				await projectService.deleteRelationship(projectId, edgeId);

				// Remove from local store
				update((state) => ({
					...state,
					edges: state.edges.filter((edge) => edge.id !== edgeId),
					selectedEdge: state.selectedEdge?.id === edgeId ? null : state.selectedEdge
				}));
			} catch (error) {
				console.error('Failed to delete relationship:', error);
				throw error;
			}
		},

		// Remove relationship edge locally (for optimistic updates)
		removeLocalRelationshipEdge(edgeId: string) {
			update((state) => ({
				...state,
				edges: state.edges.filter((edge) => edge.id !== edgeId),
				selectedEdge: state.selectedEdge?.id === edgeId ? null : state.selectedEdge
			}));
		},

		// Load relationships from backend and convert to frontend format
		async loadProjectRelationships(projectId: string): Promise<void> {
			try {
				const relationships = await projectService.getProjectRelationships(projectId);

				// Convert backend format to frontend format and add to store
				for (const rel of relationships) {
					const edgeData = {
						relationship_id: rel.relationship_id,
						source_table_id: rel.source_table_id,
						target_table_id: rel.target_table_id,
						source_field_id: rel.source_field_id,
						target_field_id: rel.target_field_id,
						relation_type: rel.relation_type
					};
					this.addLocalRelationshipEdge(edgeData);
				}

				console.log(`Loaded ${relationships.length} relationships for project ${projectId}`);
			} catch (error) {
				console.error('Failed to load project relationships:', error);
				throw error;
			}
		},

		// Select node
		selectNode(node: TableNode | null) {
			update((state) => ({
				...state,
				selectedNode: node,
				selectedEdge: null
			}));
		},

		// Select edge
		selectEdge(edge: RelationshipEdge | null) {
			update((state) => ({
				...state,
				selectedEdge: edge,
				selectedNode: null
			}));
		},

		// Update viewport
		updateViewport(viewport: { x: number; y: number; zoom: number }) {
			update((state) => ({
				...state,
				viewport
			}));
		},

		// Clear all data
		clear() {
			set(initialState);
		},

		// Force reactivity update
		forceUpdate() {
			update((state) => ({ ...state }));
		}
	};
}

export const flowStore = createFlowStore();
