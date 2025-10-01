<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteFlow, Controls, Background, MiniMap } from '@xyflow/svelte';
	import '@xyflow/svelte/dist/style.css';

	import TableNodeWrapper from './TableNodeWrapper.svelte';
	import RelationshipEdge from './RelationshipEdge.svelte';
	import MouseTracker from './MouseTracker.svelte';
	import UserCursor from '../collaboration/UserCursor.svelte';
	import CanvasHookManager from './CanvasHookManager.svelte';

	import {
		flowStore,
		type TableNode as TableNodeType,
		type RelationshipEdge as RelationshipEdgeType
	} from '$lib/stores/flow';
	import { designerStore } from '$lib/stores/designer';
	import { collaborationStore } from '$lib/stores/collaboration';
	import { projectStore } from '$lib/stores/project';
	import { authStore } from '$lib/stores/auth';
	import { projectService } from '$lib/services/project';

	// Custom node and edge types
	const nodeTypes = {
		table: TableNodeWrapper
	};

	const edgeTypes = {
		relationship: RelationshipEdge
	};

	let flowElement: HTMLElement;
	let containerRect: DOMRect | null = null;
	let canvasHookManager: CanvasHookManager;

	// Remove relationship creation state - now handled by SvelteFlow connections

	// Reactive flow data from store
	$: displayNodes = $flowStore.nodes;
	$: displayEdges = $flowStore.edges;

	// Debug edges
	$: if (displayEdges.length > 0) {
		console.log('DatabaseCanvas: displayEdges updated:', displayEdges);
	}

	// Dynamic CSS classes based on tool state
	$: canvasClasses = `database-canvas w-full h-full tool-${$designerStore.toolbar.selectedTool}`;

	// Instructions text based on current tool and state
	$: instructionText = getInstructionText($designerStore.toolbar.selectedTool);

	// Proper throttling for real-time position broadcasts
	let lastBroadcastTime = 0;
	let pendingBroadcastData: {
		nodeId: string;
		position: { x: number; y: number };
		tableName: string;
	} | null = null;
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

	function throttledBroadcastPosition(
		nodeId: string,
		position: { x: number; y: number },
		tableName: string
	) {
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

	function getInstructionText(tool: string): string | null {
		switch (tool) {
			case 'table':
				return 'Click anywhere on the canvas to place a new table';
			case 'relationship':
				return 'Drag from any field to another field to create a relationship';
			default:
				return null;
		}
	}

	// Handle node selection
	function onNodeClick(event: any) {
		// Defensive check for event structure variations
		let node: TableNodeType | null = null;

		// Try different event structures that @xyflow/svelte might use
		if (event.detail?.node) {
			node = event.detail.node as TableNodeType;
		} else if (event.node) {
			node = event.node as TableNodeType;
		} else if (event.detail?.id) {
			// Fallback: find node by ID if direct reference is missing
			const nodeId = event.detail.id;
			const foundNode = displayNodes.find((n) => n.id === nodeId);
			if (foundNode) {
				node = foundNode;
			}
		} else if (event.id) {
			// Another fallback pattern
			const nodeId = event.id;
			const foundNode = displayNodes.find((n) => n.id === nodeId);
			if (foundNode) {
				node = foundNode;
			}
		}

		// Log debug information if node is still null
		if (!node) {
			console.warn('onNodeClick: Could not extract node from event', {
				event,
				eventDetail: event.detail,
				availableNodes: displayNodes.length,
				eventKeys: Object.keys(event),
				detailKeys: event.detail ? Object.keys(event.detail) : 'no detail'
			});
			return; // Early return to prevent errors
		}

		const currentTool = $designerStore.toolbar.selectedTool;

		// For relationship tool, don't intercept table clicks - let SvelteFlow handle connections
		if (currentTool === 'relationship') {
			// Do nothing - let users drag between field handles
			return;
		}

		// Default select tool behavior
		flowStore.selectNode(node);
		designerStore.openPropertyPanel('table', node);
	}

	// Handle edge selection
	function onEdgeClick(event: any) {
		// Defensive check for event structure variations
		let edge: RelationshipEdgeType | null = null;

		// Try different event structures that @xyflow/svelte might use
		if (event.detail?.edge) {
			edge = event.detail.edge as RelationshipEdgeType;
		} else if (event.edge) {
			edge = event.edge as RelationshipEdgeType;
		} else if (event.detail?.id) {
			// Fallback: find edge by ID if direct reference is missing
			const edgeId = event.detail.id;
			const foundEdge = displayEdges.find((e) => e.id === edgeId);
			if (foundEdge) {
				edge = foundEdge;
			}
		} else if (event.id) {
			// Another fallback pattern
			const edgeId = event.id;
			const foundEdge = displayEdges.find((e) => e.id === edgeId);
			if (foundEdge) {
				edge = foundEdge;
			}
		}

		// Log debug information if edge is still null
		if (!edge) {
			console.warn('onEdgeClick: Could not extract edge from event', {
				event,
				eventDetail: event.detail,
				availableEdges: displayEdges.length,
				eventKeys: Object.keys(event),
				detailKeys: event.detail ? Object.keys(event.detail) : 'no detail'
			});
			return;
		}

		console.log('onEdgeClick: Successfully extracted edge:', edge.id);
		flowStore.selectEdge(edge);
		designerStore.openPropertyPanel('relationship', edge);
	}

	// Handle canvas click based on selected tool
	async function onPaneClick(event: any) {
		const currentTool = $designerStore.toolbar.selectedTool;

		if (currentTool === 'table') {
			// Create table when table tool is selected
			await handleTableCreation(event);
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
		// Priority order based on SvelteFlow event structure (same as onNodeDrag)
		if (event.event && event.event.clientX !== undefined && event.event.clientY !== undefined) {
			// SvelteFlow events have the MouseEvent in event.event
			screenPosition = { x: event.event.clientX, y: event.event.clientY };
		} else if (event.clientX !== undefined && event.clientY !== undefined) {
			// Direct mouse event coordinates (fallback)
			screenPosition = { x: event.clientX, y: event.clientY };
		} else if (event.detail && event.detail.event) {
			// Event wrapped in detail.event (older compatibility)
			const mouseEvent = event.detail.event;
			screenPosition = { x: mouseEvent.clientX, y: mouseEvent.clientY };
		} else if (
			event.detail &&
			typeof event.detail.clientX === 'number' &&
			typeof event.detail.clientY === 'number'
		) {
			// Some SvelteFlow versions provide clientX/clientY in detail
			screenPosition = { x: event.detail.clientX, y: event.detail.clientY };
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
				console.log('Converted coordinates:', {
					original: screenPosition,
					converted: finalPosition
				});
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
						field_id: crypto.randomUUID(),
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

			// WebSocket broadcasting is handled by the backend after successful API call
			// No need to manually broadcast here (prevents duplicate events)

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

	// Handle field-to-field connections via SvelteFlow
	async function onConnect(connection: any) {
		if (!$projectStore.currentProject) {
			console.error('No current project');
			return;
		}

		console.log('SvelteFlow connection event:', connection);

		// Parse field IDs from handle IDs (format: "tableId-fieldId-source/target")
		const sourceHandleId = connection.sourceHandle;
		const targetHandleId = connection.targetHandle;

		if (!sourceHandleId || !targetHandleId) {
			console.error('Missing handle IDs in connection:', connection);
			return;
		}

		// Extract table and field IDs from handle format: "tableId-fieldId-source"
		// UUIDs are 36 characters: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		const uuidRegex = /^([a-f0-9-]{36})-([a-f0-9-]{36})-(source|target)$/;
		const sourceMatch = sourceHandleId.match(uuidRegex);
		const targetMatch = targetHandleId.match(uuidRegex);

		if (!sourceMatch || !targetMatch) {
			console.error('Invalid handle ID format (expected UUID-UUID-type):', {
				sourceHandleId,
				targetHandleId,
				sourceMatch,
				targetMatch
			});
			return;
		}

		const sourceTableId = sourceMatch[1];
		const sourceFieldId = sourceMatch[2];
		const targetTableId = targetMatch[1];
		const targetFieldId = targetMatch[2];

		console.log('Parsed connection data:', {
			sourceTableId,
			sourceFieldId,
			targetTableId,
			targetFieldId
		});

		// Enhanced validation
		// 1. Validate that we're not connecting a field to itself
		if (sourceTableId === targetTableId && sourceFieldId === targetFieldId) {
			console.warn('Cannot connect field to itself');
			// TODO: Show user-friendly error message
			return;
		}

		// 2. Check if tables exist in current flow
		const sourceTable = displayNodes.find((node) => node.id === sourceTableId);
		const targetTable = displayNodes.find((node) => node.id === targetTableId);

		if (!sourceTable || !targetTable) {
			console.error('Source or target table not found:', { sourceTableId, targetTableId });
			return;
		}

		// 3. Check if fields exist in their respective tables
		const sourceField = sourceTable.data.fields.find((field) => field.field_id === sourceFieldId);
		const targetField = targetTable.data.fields.find((field) => field.field_id === targetFieldId);

		if (!sourceField || !targetField) {
			console.error('Source or target field not found:', {
				sourceFieldId,
				targetFieldId,
				sourceFields: sourceTable.data.fields.map((f) => f.field_id),
				targetFields: targetTable.data.fields.map((f) => f.field_id)
			});
			return;
		}

		// 4. Check if relationship already exists between these fields
		const existingRelationship = displayEdges.find(
			(edge) =>
				(edge.data.source_table_id === sourceTableId &&
					edge.data.target_table_id === targetTableId &&
					edge.data.source_field_id === sourceFieldId &&
					edge.data.target_field_id === targetFieldId) ||
				(edge.data.source_table_id === targetTableId &&
					edge.data.target_table_id === sourceTableId &&
					edge.data.source_field_id === targetFieldId &&
					edge.data.target_field_id === sourceFieldId)
		);

		if (existingRelationship) {
			console.warn('Relationship already exists between these fields');
			// TODO: Show user-friendly error message
			return;
		}

		// 5. Only allow connections when relationship tool is selected
		if ($designerStore.toolbar.selectedTool !== 'relationship') {
			console.warn('Relationship creation only allowed when relationship tool is selected');
			return;
		}

		console.log('Connection validation passed:', {
			sourceTable: sourceTable.data.name,
			sourceField: sourceField.name,
			targetTable: targetTable.data.name,
			targetField: targetField.name
		});

		try {
			// Create relationship data
			const relationshipData = {
				source_table_id: sourceTableId,
				source_field_id: sourceFieldId,
				target_table_id: targetTableId,
				target_field_id: targetFieldId,
				relation_type: 'one_to_many' as const // Default, can be changed in property panel
			};

			// Create via API
			const newRelationship = await projectService.createRelationship(
				$projectStore.currentProject.id,
				relationshipData
			);

			// Add relationship edge to local store (map backend response to frontend format)
			const edgeData = {
				relationship_id: newRelationship.relationship_id,
				source_table_id: newRelationship.source_table_id,
				target_table_id: newRelationship.target_table_id,
				source_field_id: newRelationship.source_field_id,
				target_field_id: newRelationship.target_field_id,
				relation_type: newRelationship.relation_type
			};

			flowStore.addLocalRelationshipEdge(edgeData);

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);

			console.log('Relationship created successfully:', newRelationship);
		} catch (error) {
			console.error('Failed to create relationship:', error);
			// TODO: Show error message to user
		}
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
		const tableData = displayNodes.find((n) => n.id === node.id)?.data;
		const tableName = tableData?.name || 'Unknown Table';

		// Broadcast position in real-time (throttled)
		throttledBroadcastPosition(node.id, node.position, tableName);

		// Broadcast cursor position during drag
		// Mouse coordinates are available in event.event (MouseEvent)
		if (canvasHookManager && event.event) {
			const mouseEvent = event.event;
			if (mouseEvent.clientX !== undefined && mouseEvent.clientY !== undefined) {
				try {
					// Convert screen coordinates to flow coordinates
					const flowCoords = canvasHookManager.convertScreenToFlow(
						mouseEvent.clientX,
						mouseEvent.clientY
					);

					// Broadcast cursor position
					if (isFinite(flowCoords.x) && isFinite(flowCoords.y)) {
						collaborationStore.sendCursorPosition(flowCoords.x, flowCoords.y);
					}
				} catch (error) {
					console.warn('Error broadcasting cursor during drag:', error);
				}
			}
		}
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
			const tableData = displayNodes.find((n) => n.id === node.id)?.data;
			collaborationStore.sendSchemaEvent('table_moved', {
				table_id: node.id,
				name: tableData?.name || 'Unknown Table',
				x: position.x,
				y: position.y
			});

			// Update position in backend and local store
			console.log('SAVE DEBUG: Calling updateTablePosition...');
			await flowStore.updateTablePosition($projectStore.currentProject.id, node.id, position);
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

		// Helper function to check if user is currently editing text
		function isUserEditingText(): boolean {
			const activeElement = document.activeElement;
			if (!activeElement) return false;

			// Check if the active element is an input field, textarea, or content editable
			const tagName = activeElement.tagName.toLowerCase();
			const isTextInput = tagName === 'input' || tagName === 'textarea';
			const isContentEditable = activeElement.getAttribute('contenteditable') === 'true';

			// Also check if it's an input with text type (exclude buttons, checkboxes, etc.)
			if (tagName === 'input') {
				const inputType = (activeElement as HTMLInputElement).type;
				const textInputTypes = ['text', 'email', 'password', 'search', 'tel', 'url'];
				return textInputTypes.includes(inputType);
			}

			return isTextInput || isContentEditable;
		}

		// Set up keyboard shortcuts
		async function handleKeydown(event: KeyboardEvent) {
			if (!$projectStore.currentProject) return;

			if (event.key === 'Escape') {
				// Clear any selections
				flowStore.selectNode(null);
				flowStore.selectEdge(null);
				designerStore.closePropertyPanel();
			}

			if (event.key === 'Delete' || event.key === 'Backspace') {
				// Don't delete tables if user is editing text in an input field
				if (isUserEditingText()) {
					return; // Allow normal text editing behavior
				}

				if ($flowStore.selectedNode) {
					// Capture selectedNode data before any operations to prevent null reference errors
					const nodeToDelete = $flowStore.selectedNode;

					// Validate that we have all required data
					if (!nodeToDelete || !nodeToDelete.data || !nodeToDelete.id) {
						console.error('Cannot delete: invalid node data', nodeToDelete);
						return;
					}

					// Capture the data we need before any state changes
					const tableData = {
						id: nodeToDelete.id,
						name: nodeToDelete.data.name
					};

					console.log('🗑️ Deleting table:', tableData);

					try {
						// Delete table via API - backend will handle WebSocket notifications
						await flowStore.removeTableNode($projectStore.currentProject.id, tableData.id);
						console.log('✅ Table deletion completed');

						// Auto-save canvas data
						const canvasData = flowStore.getCurrentCanvasData();
						projectStore.autoSaveCanvasData(canvasData);
					} catch (error) {
						console.error('❌ Failed to delete table:', error);
					}
				} else if ($flowStore.selectedEdge) {
					// Capture selectedEdge data before any operations to prevent null reference errors
					const edgeToDelete = $flowStore.selectedEdge;

					// Validate that we have all required data
					if (!edgeToDelete || !edgeToDelete.id) {
						console.error('Cannot delete: invalid edge data', edgeToDelete);
						return;
					}

					console.log('🗑️ Deleting relationship:', edgeToDelete.id);

					try {
						// Delete relationship via API
						await projectService.deleteRelationship(
							$projectStore.currentProject.id,
							edgeToDelete.id
						);

						// Remove from local store
						flowStore.removeLocalRelationshipEdge(edgeToDelete.id);

						// WebSocket broadcasting is now handled by the backend after successful API call

						// Auto-save canvas data
						const canvasData = flowStore.getCurrentCanvasData();
						projectStore.autoSaveCanvasData(canvasData);

						console.log('✅ Relationship deletion completed');
					} catch (error) {
						console.error('❌ Failed to delete relationship:', error);
						// TODO: Show error message to user
					}
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
		{edgeTypes}
		fitView
		snapGrid={[$designerStore.gridSize, $designerStore.gridSize]}
		onnodeclick={onNodeClick}
		onedgeclick={onEdgeClick}
		onpaneclick={onPaneClick}
		onmove={onMove}
		onnodedrag={onNodeDrag}
		onnodedragstop={onNodeDragStop}
		onconnect={onConnect}
	>
		<!-- Hook manager component - provides hook access to parent -->
		<CanvasHookManager bind:this={canvasHookManager} />

		<!-- Mouse tracking component - must be inside SvelteFlow for hook access -->
		<MouseTracker />

		<!-- Live Cursors - inside SvelteFlow context for hook access -->
		{#each $collaborationStore.connectedUsers as user}
			{#if user.cursor && user.id !== $authStore.user?.id}
				<UserCursor {user} />
			{/if}
		{/each}

		<!-- Background with grid -->
		<Background gap={$designerStore.gridSize} />

		<!-- Controls for zoom/pan -->
		<Controls />

		<!-- Minimap -->
		{#if $designerStore.showMinimap}
			<MiniMap nodeColor="#3b82f6" maskColor="rgba(0, 0, 0, 0.1)" position="bottom-right" />
		{/if}
	</SvelteFlow>

	<!-- Tool Instructions Overlay -->
	{#if instructionText && $designerStore.toolbar.selectedTool !== 'select'}
		<div class="tool-instructions">
			{instructionText}
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

	/* Table highlighting removed - relationships now created via field connections */

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
