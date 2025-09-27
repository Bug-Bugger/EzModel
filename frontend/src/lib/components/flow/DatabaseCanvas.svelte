<script lang="ts">
	import { onMount } from 'svelte';
	import { SvelteFlow, Controls, Background, MiniMap } from '@xyflow/svelte';
	import '@xyflow/svelte/dist/style.css';

	import TableNode from './TableNode.svelte';
		import MouseTracker from './MouseTracker.svelte';
	import UserCursor from '../collaboration/UserCursor.svelte';

	import { flowStore, type TableNode as TableNodeType, type RelationshipEdge as RelationshipEdgeType } from '$lib/stores/flow';
	import { designerStore } from '$lib/stores/designer';
	import { collaborationStore } from '$lib/stores/collaboration';

	// Custom node and edge types
	const nodeTypes = {
		table: TableNode
	};

	// TODO: Fix edge types compatibility with @xyflow/svelte v1.3.1
	// const edgeTypes = {
	// 	relationship: RelationshipEdge
	// };

	let flowElement: HTMLElement;
	let containerRect: DOMRect | null = null;

	// Reactive flow data
	$: nodes = $flowStore.nodes;
	$: edges = $flowStore.edges;

	// Handle node selection
	function onNodeClick(event: any) {
		const node = event.detail.node as TableNodeType;
		flowStore.selectNode(node);
		designerStore.openPropertyPanel('table', node);
	}

	// Handle edge selection
	function onEdgeClick(event: any) {
		const edge = event.detail.edge as RelationshipEdgeType;
		flowStore.selectEdge(edge);
		designerStore.openPropertyPanel('relationship', edge);
	}

	// Handle canvas click (deselect)
	function onPaneClick() {
		flowStore.selectNode(null);
		flowStore.selectEdge(null);
		designerStore.closePropertyPanel();
	}





	// Update container bounds when element or viewport changes
	function updateContainerBounds() {
		if (flowElement) {
			containerRect = flowElement.getBoundingClientRect();
		}
	}

	// Handle viewport changes
	function onMove(event: any) {
		const viewport = event.detail.viewport;
		flowStore.updateViewport(viewport);
		designerStore.setZoom(viewport.zoom);

		// Update container bounds when viewport changes (pan/zoom)
		updateContainerBounds();
	}

	// Handle canvas double-click to create table
	function onPaneDoubleClick(event: any) {
		if ($designerStore.toolbar.selectedTool === 'table' || $designerStore.toolbar.isCreatingTable) {
			const position = event.detail.position;

			// Create new table
			const newTable = {
				id: crypto.randomUUID(),
				name: 'New Table',
				fields: [
					{
						id: crypto.randomUUID(),
						name: 'id',
						type: 'UUID',
						isPrimary: true,
						isForeign: false,
						isRequired: true,
						isUnique: true
					}
				]
			};

			const tableNode = flowStore.addTableNode(newTable, position);

			// Send collaboration event
			collaborationStore.sendSchemaEvent('table_create', newTable);

			// Select the new table for editing
			flowStore.selectNode(tableNode);
			designerStore.openPropertyPanel('table', tableNode);
			designerStore.finishTableCreation();
		}
	}

	// Handle node drag end to save position
	function onNodeDragStop(event: any) {
		const node = event.detail.node;
		flowStore.updateTableNode(node.id, { position: node.position });

		// Send collaboration event
		collaborationStore.sendSchemaEvent('table_update', {
			id: node.id,
			position: node.position
		});
	}


	onMount(() => {
		// Initialize container bounds
		updateContainerBounds();


		// Update bounds on window resize
		function handleResize() {
			updateContainerBounds();
		}

		// Set up keyboard shortcuts
		function handleKeydown(event: KeyboardEvent) {
			if (event.key === 'Delete' || event.key === 'Backspace') {
				if ($flowStore.selectedNode) {
					flowStore.removeTableNode($flowStore.selectedNode.id);
					collaborationStore.sendSchemaEvent('table_delete', {
						id: $flowStore.selectedNode.id,
						name: $flowStore.selectedNode.data.name
					});
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

		};
	});
</script>

<div
	class="database-canvas w-full h-full"
	bind:this={flowElement}
	role="application"
	aria-label="Database schema designer canvas"
>
	<SvelteFlow
		{nodes}
		{edges}
		{nodeTypes}
		fitView
		snapGrid={[$designerStore.gridSize, $designerStore.gridSize]}
		onnodeclick={onNodeClick}
		onedgeclick={onEdgeClick}
		onpaneclick={onPaneClick}
		onmove={onMove}
		ondblclick={onPaneDoubleClick}
		onnodedragstop={onNodeDragStop}
	>
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
</style>