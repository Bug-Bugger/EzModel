<script lang="ts">
	import { onMount } from 'svelte';
	import {
		SvelteFlow,
		Controls,
		Background,
		MiniMap
	} from '@xyflow/svelte';
	import '@xyflow/svelte/dist/style.css';

	import TableNodeWrapper from './TableNodeWrapper.svelte';
	import MouseTracker from './MouseTracker.svelte';
	import UserCursor from '../collaboration/UserCursor.svelte';
	import CanvasHookManager from './CanvasHookManager.svelte';

	import { flowStore, type TableNode as TableNodeType, type RelationshipEdge as RelationshipEdgeType } from '$lib/stores/flow';
	import { designerStore } from '$lib/stores/designer';
	import { collaborationStore } from '$lib/stores/collaboration';
	import { projectStore } from '$lib/stores/project';
	import { projectService } from '$lib/services/project';

	// Custom node and edge types
	const nodeTypes = {
		table: TableNodeWrapper
	};

	// TODO: Fix edge types compatibility with @xyflow/svelte v1.3.1
	// const edgeTypes = {
	// 	relationship: RelationshipEdge
	// };

	let flowElement: HTMLElement;
	let containerRect: DOMRect | null = null;
	let canvasHookManager: CanvasHookManager;

	// Relationship creation state
	let relationshipCreation = {
		isActive: false,
		firstTableId: null as string | null,
		firstTableName: null as string | null
	};

	// Reactive flow data from store
	$: displayNodes = $flowStore.nodes;
	$: displayEdges = $flowStore.edges;

	// Dynamic CSS classes based on tool state
	$: canvasClasses = `database-canvas w-full h-full tool-${$designerStore.toolbar.selectedTool}${relationshipCreation.isActive ? ' relationship-active' : ''}`;

	// Instructions text based on current tool and state
	$: instructionText = getInstructionText($designerStore.toolbar.selectedTool, relationshipCreation.isActive, relationshipCreation.firstTableName);

	// Proper throttling for real-time position broadcasts
	let lastBroadcastTime = 0;
	let pendingBroadcastData: { nodeId: string; position: { x: number; y: number }; tableName: string } | null = null;
	let broadcastTimer: ReturnType<typeof setTimeout> | null = null;
	const BROADCAST_THROTTLE_MS = 50; // ~20 FPS for smooth collaboration

	function broadcastNow() {
		if (pendingBroadcastData) {
			// Send collaboration event with proper TablePayload structure
			collaborationStore.sendSchemaEvent('table_updated', {
				table_id: pendingBroadcastData.nodeId,
				name: pendingBroadcastData.tableName,
				x: pendingBroadcastData.position.x,
				y: pendingBroadcastData.position.y
			});

			lastBroadcastTime = Date.now();
			pendingBroadcastData = null;
			broadcastTimer = null;
		}
	}

	function throttledBroadcastPosition(nodeId: string, position: { x: number; y: number }, tableName: string) {
		const now = Date.now();
		const timeSinceLastBroadcast = now - lastBroadcastTime;

		// Always store the latest position data
		pendingBroadcastData = { nodeId, position, tableName };

		if (timeSinceLastBroadcast >= BROADCAST_THROTTLE_MS) {
			// Enough time has passed, broadcast immediately
			broadcastNow();
		} else if (!broadcastTimer) {
			// Schedule broadcast for the remaining time
			const delay = BROADCAST_THROTTLE_MS - timeSinceLastBroadcast;
			broadcastTimer = setTimeout(broadcastNow, delay);
		}
		// If timer is already running, we just updated pendingBroadcastData with latest position
	}

	function getInstructionText(tool: string, relationshipActive: boolean, firstTableName: string | null): string | null {
		switch (tool) {
			case 'table':
				return 'Click anywhere on the canvas to place a new table';
			case 'relationship':
				if (relationshipActive && firstTableName) {
					return `Click on another table to create a relationship from "${firstTableName}"`;
				}
				return 'Click on a table to start creating a relationship';
			default:
				return null;
		}
	}

	// Handle node selection
	function onNodeClick(event: any) {
		const node = event.detail.node as TableNodeType;
		const currentTool = $designerStore.toolbar.selectedTool;

		if (currentTool === 'relationship') {
			// Handle relationship creation workflow
			handleRelationshipNodeClick(node);
		} else {
			// Default select tool behavior
			flowStore.selectNode(node);
			designerStore.openPropertyPanel('table', node);
		}
	}

	// Handle edge selection
	function onEdgeClick(event: any) {
		const edge = event.detail.edge as RelationshipEdgeType;
		flowStore.selectEdge(edge);
		designerStore.openPropertyPanel('relationship', edge);
	}


	// Handle canvas click based on selected tool
	async function onPaneClick(event: any) {
		const currentTool = $designerStore.toolbar.selectedTool;

		if (currentTool === 'table') {
			// Create table when table tool is selected
			await handleTableCreation(event);
		} else if (currentTool === 'relationship') {
			// Handle relationship creation workflow
			handleRelationshipClick(event);
		} else {
			// Default select tool behavior - deselect all
			flowStore.selectNode(null);
			flowStore.selectEdge(null);
			designerStore.closePropertyPanel();
		}
	}

	// Selection change handling is now done via onNodeClick/onEdgeClick events





	// Update container bounds when element or viewport changes
	function updateContainerBounds() {
		if (flowElement) {
			containerRect = flowElement.getBoundingClientRect();
		}
	}

	// Handle viewport changes
	function onMove(event: any) {
		const viewport = event.detail?.viewport;

		// Ensure viewport is valid before processing
		if (viewport && typeof viewport.zoom === 'number' && isFinite(viewport.zoom)) {
			flowStore.updateViewport(viewport);
			designerStore.setZoom(viewport.zoom);

			// Update container bounds when viewport changes (pan/zoom)
			updateContainerBounds();
		}
	}

	// Handle table creation (extracted from double-click)
	async function handleTableCreation(event: any) {
		if (!$projectStore.currentProject) {
			console.error('No current project');
			return;
		}

		// Check if flow is ready for table creation
		if (!flowElement || !canvasHookManager) {
			console.warn('Flow not ready for table creation, skipping');
			return;
		}

		// Check if nodes are initialized
		if (!canvasHookManager.getNodesInitialized()) {
			console.warn('Nodes not yet initialized, skipping table creation');
			return;
		}

		// Get mouse coordinates from the event
		let screenPosition = { x: 0, y: 0 };

		// Try to extract coordinates from various event properties
		if (event.detail && typeof event.detail.clientX === 'number' && typeof event.detail.clientY === 'number') {
			// Some SvelteFlow versions provide clientX/clientY in detail
			screenPosition = { x: event.detail.clientX, y: event.detail.clientY };
		} else if (event.clientX !== undefined && event.clientY !== undefined) {
			// Standard mouse event coordinates
			screenPosition = { x: event.clientX, y: event.clientY };
		} else if (event.detail && event.detail.event) {
			// Event wrapped in detail.event
			const mouseEvent = event.detail.event;
			screenPosition = { x: mouseEvent.clientX, y: mouseEvent.clientY };
		} else {
			// Fallback: use center of canvas if available
			if (containerRect) {
				screenPosition = {
					x: containerRect.left + containerRect.width / 2,
					y: containerRect.top + containerRect.height / 2
				};
			}
		}

		// Debug logging to understand coordinate issues
		console.log('Table creation clicked:', {
			eventDetail: event.detail,
			screenPosition: screenPosition,
			containerRect: containerRect,
			canvasHookManager: !!canvasHookManager,
			flowElement: !!flowElement
		});

		// Convert screen coordinates to flow coordinates
		let finalPosition = screenPosition;
		if (canvasHookManager && screenPosition) {
			try {
				// Use the hook manager for proper coordinate conversion
				const converted = canvasHookManager.convertScreenToFlow(screenPosition.x, screenPosition.y);
				finalPosition = converted;
				console.log('Converted coordinates:', { original: screenPosition, converted: finalPosition });
			} catch (error) {
				console.warn('Coordinate conversion failed, using screen position:', error);
				// Fallback to screen position if conversion fails
				finalPosition = screenPosition;
			}
		}

		try {
			// Create new table data (without ID, backend will generate)
			const newTableData = {
				name: 'New Table',
				fields: [
					{
						id: crypto.randomUUID(),
						table_id: '', // Will be set by backend
						name: 'id',
						data_type: 'UUID',
						is_primary_key: true,
						is_nullable: false,
						default_value: '',
						position: 0,
						created_at: new Date().toISOString(),
						updated_at: new Date().toISOString()
					}
				]
			};

			// Create table via API-integrated store method
			const tableNode = await flowStore.addTableNode(
				$projectStore.currentProject.id,
				newTableData,
				finalPosition
			);

			// Send collaboration event with essential table data
			collaborationStore.sendSchemaEvent('table_created', {
				id: tableNode.id,
				name: tableNode.data.name,
				pos_x: finalPosition.x,
				pos_y: finalPosition.y,
				fields: tableNode.data.fields
			});

			// Select the new table for editing
			flowStore.selectNode(tableNode);
			designerStore.openPropertyPanel('table', tableNode);

			// Auto-save canvas data
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);

			// Switch back to select tool after creating table
			designerStore.selectTool('select');
		} catch (error) {
			console.error('Failed to create table:', error);
			// TODO: Show error message to user
		}
	}

	// Handle clicking on canvas when relationship tool is selected
	function handleRelationshipClick(event: any) {
		// Cancel relationship creation if user clicks on empty canvas
		if (relationshipCreation.isActive) {
			cancelRelationshipCreation();
		}
	}

	// Handle clicking on a table node when relationship tool is selected
	async function handleRelationshipNodeClick(node: TableNodeType) {
		if (!$projectStore.currentProject) {
			console.error('No current project');
			return;
		}

		if (!relationshipCreation.isActive) {
			// First click - select the source table
			relationshipCreation.isActive = true;
			relationshipCreation.firstTableId = node.id;
			relationshipCreation.firstTableName = node.data.name;

			console.log(`Started relationship from table: ${node.data.name}`);
			// TODO: Add visual feedback to highlight the selected table
		} else {
			// Second click - create the relationship
			if (node.id === relationshipCreation.firstTableId) {
				// User clicked the same table - cancel creation
				console.warn('Cannot create relationship to the same table');
				cancelRelationshipCreation();
				return;
			}

			await createRelationship(relationshipCreation.firstTableId!, node.id);
		}
	}

	// Create relationship between two tables
	async function createRelationship(fromTableId: string, toTableId: string) {
		if (!$projectStore.currentProject) return;

		try {
			// For now, create a basic relationship
			// TODO: Let user choose field mappings and relationship type
			const relationshipData = {
				name: `${relationshipCreation.firstTableName}_to_${displayNodes.find(n => n.id === toTableId)?.data.name}`,
				from_table_id: fromTableId,
				to_table_id: toTableId,
				from_field_id: 'id', // Default to primary key
				to_field_id: 'id',   // Default to primary key
				relationship_type: 'one-to-many' as const
			};

			// Create relationship via API
			const relationship = await projectService.createRelationship(
				$projectStore.currentProject.id,
				relationshipData
			);

			// Add to flow store
			flowStore.addRelationshipEdge({
				id: relationship.id,
				fromTable: fromTableId,
				toTable: toTableId,
				fromField: relationshipData.from_field_id,
				toField: relationshipData.to_field_id,
				type: relationshipData.relationship_type
			});

			// Send collaboration event
			collaborationStore.sendSchemaEvent('relationship_create', relationship);

			// Auto-save canvas data
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);

			console.log('Relationship created successfully');
		} catch (error) {
			console.error('Failed to create relationship:', error);
			// TODO: Show error message to user
		} finally {
			// Reset state and switch back to select tool
			cancelRelationshipCreation();
			designerStore.selectTool('select');
		}
	}

	// Cancel relationship creation
	function cancelRelationshipCreation() {
		relationshipCreation.isActive = false;
		relationshipCreation.firstTableId = null;
		relationshipCreation.firstTableName = null;
		console.log('Relationship creation cancelled');
	}

	// Handle real-time node dragging for collaboration
	function onNodeDrag(event: any) {
		if (!$projectStore.currentProject) {
			return;
		}

		// SvelteFlow drag events use targetNode structure
		const node = event.targetNode;

		if (!node || !node.id || !node.position) {
			return;
		}

		// Get table name for the broadcast
		const tableData = displayNodes.find(n => n.id === node.id)?.data;
		const tableName = tableData?.name || 'Unknown Table';

		// Broadcast position in real-time (throttled)
		throttledBroadcastPosition(node.id, node.position, tableName);
	}

	// Handle node drag end to save position
	async function onNodeDragStop(event: any) {
		if (!$projectStore.currentProject) {
			console.error('No current project');
			return;
		}

		// SvelteFlow drag events use targetNode structure
		const node = event.targetNode;

		if (!node) {
			console.error('targetNode is missing in drag event:', event);
			return;
		}

		if (!node.id) {
			console.error('targetNode.id is missing:', { event, targetNode: node });
			return;
		}

		if (!node.position) {
			console.error('targetNode.position is missing:', { event, targetNode: node });
			return;
		}

		const position = node.position;
		console.log('Drag successful - updating position for table:', {
			id: node.id,
			position: position
		});

		try {
			console.log('SAVE DEBUG: Starting position save process...');

			// Clear any pending broadcast timer and send final position
			if (broadcastTimer) {
				clearTimeout(broadcastTimer);
				broadcastTimer = null;
			}

			// Send final position broadcast for real-time updates
			if (pendingBroadcastData) {
				broadcastNow();
			}

			// Send separate "table moved" event for activity logging
			const tableData = displayNodes.find(n => n.id === node.id)?.data;
			collaborationStore.sendSchemaEvent('table_moved', {
				table_id: node.id,
				name: tableData?.name || 'Unknown Table',
				x: position.x,
				y: position.y
			});

			// Update position in backend and local store
			console.log('SAVE DEBUG: Calling updateTablePosition...');
			await flowStore.updateTablePosition(
				$projectStore.currentProject.id,
				node.id,
				position
			);
			console.log('SAVE DEBUG: updateTablePosition completed');

			// Final position already sent by broadcastNow() above

			// Auto-save canvas data
			console.log('SAVE DEBUG: Getting canvas data...');
			const canvasData = flowStore.getCurrentCanvasData();
			console.log('SAVE DEBUG: Canvas data to save:', canvasData);

			console.log('SAVE DEBUG: Calling autoSaveCanvasData...');
			projectStore.autoSaveCanvasData(canvasData);
			console.log('SAVE DEBUG: autoSaveCanvasData called (debounced)');
		} catch (error) {
			console.error('Failed to update table position:', error);
			// TODO: Show error message to user
		}
	}


	onMount(() => {
		// Initialize container bounds
		updateContainerBounds();


		// Update bounds on window resize
		function handleResize() {
			updateContainerBounds();
		}

		// Set up keyboard shortcuts
		async function handleKeydown(event: KeyboardEvent) {
			if (!$projectStore.currentProject) return;

			if (event.key === 'Escape') {
				// Cancel relationship creation if active
				if (relationshipCreation.isActive) {
					cancelRelationshipCreation();
					event.preventDefault();
					return;
				}
			}

			if (event.key === 'Delete' || event.key === 'Backspace') {
				if ($flowStore.selectedNode) {
					try {
						await flowStore.removeTableNode(
							$projectStore.currentProject.id,
							$flowStore.selectedNode.id
						);
						collaborationStore.sendSchemaEvent('table_deleted', {
							id: $flowStore.selectedNode.id,
							name: $flowStore.selectedNode.data.name
						});

						// Auto-save canvas data
						const canvasData = flowStore.getCurrentCanvasData();
						projectStore.autoSaveCanvasData(canvasData);
					} catch (error) {
						console.error('Failed to delete table:', error);
					}
				} else if ($flowStore.selectedEdge) {
					flowStore.removeRelationshipEdge($flowStore.selectedEdge.id);
					collaborationStore.sendSchemaEvent('relationship_delete', {
						id: $flowStore.selectedEdge.id
					});
				}
			}
		}

		window.addEventListener('keydown', handleKeydown);
		window.addEventListener('resize', handleResize);

		return () => {
			window.removeEventListener('keydown', handleKeydown);
			window.removeEventListener('resize', handleResize);

			// Clean up throttle timer and reset state
			if (broadcastTimer) {
				clearTimeout(broadcastTimer);
				broadcastTimer = null;
			}
			pendingBroadcastData = null;
			lastBroadcastTime = 0;
		};
	});
</script>

<div
	class={canvasClasses}
	bind:this={flowElement}
	role="application"
	aria-label="Database schema designer canvas"
>
	<SvelteFlow
		nodes={displayNodes}
		edges={displayEdges}
		{nodeTypes}
		fitView
		snapGrid={[$designerStore.gridSize, $designerStore.gridSize]}
		onnodeclick={onNodeClick}
		onedgeclick={onEdgeClick}
		onpaneclick={onPaneClick}
		onmove={onMove}
		onnodedrag={onNodeDrag}
		onnodedragstop={onNodeDragStop}
	>
		<!-- Hook manager component - provides hook access to parent -->
		<CanvasHookManager bind:this={canvasHookManager} />

		<!-- Mouse tracking component - must be inside SvelteFlow for hook access -->
		<MouseTracker />

		<!-- Live Cursors - inside SvelteFlow context for hook access -->
		{#each $collaborationStore.connectedUsers as user}
			{#if user.cursor}
				<UserCursor {user} />
			{/if}
		{/each}

		<!-- Background with grid -->
		<Background
			gap={$designerStore.gridSize}
		/>

		<!-- Controls for zoom/pan -->
		<Controls />

		<!-- Minimap -->
		{#if $designerStore.showMinimap}
			<MiniMap
				nodeColor="#3b82f6"
				maskColor="rgba(0, 0, 0, 0.1)"
				position="bottom-right"
			/>
		{/if}
	</SvelteFlow>

	<!-- Tool Instructions Overlay -->
	{#if instructionText && ($designerStore.toolbar.selectedTool !== 'select')}
		<div class="tool-instructions">
			{instructionText}
			{#if $designerStore.toolbar.selectedTool === 'relationship' && relationshipCreation.isActive}
				<div class="text-xs mt-2 opacity-75">Press Escape to cancel</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.database-canvas :global(.svelte-flow__node-table) {
		background: white;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		min-width: 200px;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
	}

	.database-canvas :global(.svelte-flow__node-table.selected) {
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.database-canvas :global(.svelte-flow__edge-relationship) {
		stroke: #64748b;
		stroke-width: 2;
	}

	.database-canvas :global(.svelte-flow__edge-relationship.selected) {
		stroke: #3b82f6;
		stroke-width: 3;
	}

	/* Tool-specific cursor styles */
	.database-canvas.tool-select {
		cursor: default;
	}

	.database-canvas.tool-table {
		cursor: crosshair;
	}

	.database-canvas.tool-relationship {
		cursor: crosshair;
	}

	.database-canvas.tool-relationship.relationship-active {
		cursor: copy;
	}

	/* Table highlighting during relationship creation */
	.database-canvas :global(.svelte-flow__node-table.relationship-source) {
		border-color: #10b981;
		box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2);
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% {
			opacity: 1;
		}
		50% {
			opacity: 0.8;
		}
	}

	/* Instructions overlay */
	.tool-instructions {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(0, 0, 0, 0.8);
		color: white;
		padding: 1rem 1.5rem;
		border-radius: 8px;
		pointer-events: none;
		z-index: 1000;
		font-size: 0.875rem;
		text-align: center;
		max-width: 300px;
	}
</style>