import { writable } from 'svelte/store';
import type { TableField } from './flow';

export interface PropertyPanelState {
	isOpen: boolean;
	type: 'table' | 'field' | 'relationship' | null;
	target: any;
}

export interface ToolbarState {
	selectedTool: 'select' | 'table' | 'relationship';
	isCreatingTable: boolean;
	isCreatingRelationship: boolean;
}

interface DesignerState {
	propertyPanel: PropertyPanelState;
	toolbar: ToolbarState;
	isExporting: boolean;
	exportFormat: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver';
	showGrid: boolean;
	snapToGrid: boolean;
	gridSize: number;
	zoom: number;
	showMinimap: boolean;
}

function createDesignerStore() {
	const initialState: DesignerState = {
		propertyPanel: {
			isOpen: false,
			type: null,
			target: null
		},
		toolbar: {
			selectedTool: 'select',
			isCreatingTable: false,
			isCreatingRelationship: false
		},
		isExporting: false,
		exportFormat: 'postgresql',
		showGrid: true,
		snapToGrid: true,
		gridSize: 20,
		zoom: 1,
		showMinimap: true
	};

	const { subscribe, set, update } = writable(initialState);

	return {
		subscribe,

		// Property Panel Management
		openPropertyPanel(type: 'table' | 'field' | 'relationship', target: any) {
			update(state => ({
				...state,
				propertyPanel: {
					isOpen: true,
					type,
					target
				}
			}));
		},

		closePropertyPanel() {
			update(state => ({
				...state,
				propertyPanel: {
					...state.propertyPanel,
					isOpen: false
				}
			}));
		},

		updatePropertyTarget(target: any) {
			update(state => ({
				...state,
				propertyPanel: {
					...state.propertyPanel,
					target
				}
			}));
		},

		// Toolbar Management
		selectTool(tool: 'select' | 'table' | 'relationship') {
			update(state => ({
				...state,
				toolbar: {
					...state.toolbar,
					selectedTool: tool,
					isCreatingTable: tool === 'table',
					isCreatingRelationship: tool === 'relationship'
				}
			}));
		},

		startTableCreation() {
			update(state => ({
				...state,
				toolbar: {
					...state.toolbar,
					selectedTool: 'table',
					isCreatingTable: true
				}
			}));
		},

		finishTableCreation() {
			update(state => ({
				...state,
				toolbar: {
					...state.toolbar,
					selectedTool: 'select',
					isCreatingTable: false
				}
			}));
		},

		startRelationshipCreation() {
			update(state => ({
				...state,
				toolbar: {
					...state.toolbar,
					selectedTool: 'relationship',
					isCreatingRelationship: true
				}
			}));
		},

		finishRelationshipCreation() {
			update(state => ({
				...state,
				toolbar: {
					...state.toolbar,
					selectedTool: 'select',
					isCreatingRelationship: false
				}
			}));
		},

		// Export Management
		startExport(format: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver') {
			update(state => ({
				...state,
				isExporting: true,
				exportFormat: format
			}));
		},

		finishExport() {
			update(state => ({
				...state,
				isExporting: false
			}));
		},

		// Canvas Settings
		toggleGrid() {
			update(state => ({
				...state,
				showGrid: !state.showGrid
			}));
		},

		toggleSnapToGrid() {
			update(state => ({
				...state,
				snapToGrid: !state.snapToGrid
			}));
		},

		setGridSize(size: number) {
			update(state => ({
				...state,
				gridSize: size
			}));
		},

		setZoom(zoom: number) {
			update(state => ({
				...state,
				zoom
			}));
		},

		toggleMinimap() {
			update(state => ({
				...state,
				showMinimap: !state.showMinimap
			}));
		},

		// Reset all state
		reset() {
			set(initialState);
		}
	};
}

export const designerStore = createDesignerStore();