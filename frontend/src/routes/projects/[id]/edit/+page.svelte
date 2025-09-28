<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { SvelteFlow, Controls, Background, MiniMap } from '@xyflow/svelte';
	import '@xyflow/svelte/dist/style.css';

	import DatabaseCanvas from '$lib/components/flow/DatabaseCanvas.svelte';
	import Toolbar from '$lib/components/flow/Toolbar.svelte';
	import PropertyPanel from '$lib/components/flow/PropertyPanel.svelte';
	import CollaborationStatus from '$lib/components/collaboration/CollaborationStatus.svelte';
	import PresenceList from '$lib/components/collaboration/PresenceList.svelte';
	import ActivityFeed from '$lib/components/collaboration/ActivityFeed.svelte';
	
	import { projectStore } from '$lib/stores/project.js';
	import { collaborationStore } from '$lib/stores/collaboration.js';
	import { flowStore } from '$lib/stores/flow.js';
	import { designerStore } from '$lib/stores/designer.js';
	import { authStore } from '$lib/stores/auth.js';
	import { projectService } from '$lib/services/project.js';

	const projectId = $page.params.id;

	let canvasContainer: HTMLElement;
	let showLeftSidebar = true;
	let showRightSidebar = true;

	onMount(async () => {
		// Load project data
		if (projectId) {
			// Ensure project loads first and canvas_data is available
			await projectStore.loadProject(projectId);

			// Wait for project store to be fully populated (reactive store might need time)
			let retryCount = 0;
			while (!$projectStore.currentProject && retryCount < 10) {
				await new Promise(resolve => setTimeout(resolve, 50));
				retryCount++;
			}

			if (!$projectStore.currentProject) {
				console.error('Failed to load project after retries');
				return;
			}

			// Initialize WebSocket connection for collaboration
			await collaborationStore.connect(projectId);

			// Load tables and relationships from backend
			await loadProjectData(projectId);
		}
	});

	async function loadProjectData(projectId: string) {
		try {
			// Load tables and relationships from backend
			const [tables, relationships] = await Promise.all([
				projectService.getProjectTables(projectId),
				projectService.getProjectRelationships(projectId)
			]);

			// Parse existing canvas data to get positioning information
			let savedPositions: Record<string, { x: number; y: number }> = {};
			console.log('DEBUG: Raw canvas_data from project:', $projectStore.currentProject?.canvas_data);

			if ($projectStore.currentProject?.canvas_data) {
				try {
					const canvasData = JSON.parse($projectStore.currentProject.canvas_data);
					console.log('DEBUG: Parsed canvas data object:', canvasData);

					// Extract positions from saved nodes
					if (canvasData.nodes) {
						console.log('DEBUG: Found nodes in canvas data:', canvasData.nodes.length);
						savedPositions = canvasData.nodes.reduce((acc: any, node: any) => {
							console.log(`DEBUG: Extracting position for node ${node.id}:`, node.position);
							acc[node.id] = node.position;
							return acc;
						}, {});
						console.log('DEBUG: Extracted positions:', savedPositions);
					} else {
						console.log('DEBUG: No nodes found in canvas data');
					}
				} catch (error) {
					console.warn('Failed to parse saved canvas data:', error);
				}
			} else {
				console.log('DEBUG: No canvas_data found in project');
			}

			// Clear existing flow state
			flowStore.clear();

			// Reconstruct table nodes from backend data with field loading
			console.log('DEBUG: Starting table reconstruction...');
			for (const table of tables) {
				const savedPosition = savedPositions[table.id];

				// For debugging: use fixed position if no saved position found
				const position = savedPosition || {
					x: 100 + (Object.keys(savedPositions).length * 200), // Fixed position based on index
					y: 100
				};

				console.log(`DEBUG: Reconstructing table "${table.name}" (${table.id}):`, {
					tableFromBackend: table,
					savedPosition,
					finalPosition: position,
					usingRandomPosition: !savedPosition
				});

				// Load field data for this table since backend doesn't include it
				let tableFields = [];
				try {
					tableFields = await projectService.getTableFields(projectId, table.id);
					console.log(`DEBUG: Loaded ${tableFields.length} fields for table "${table.name}"`);
				} catch (error) {
					console.warn(`Failed to load fields for table "${table.name}":`, error);
					// If no fields exist yet, create a default ID field
					tableFields = [];
				}

				// Convert backend table to frontend table node format
				const tableData = {
					id: table.id,
					name: table.name,
					fields: tableFields
				};

				flowStore.addLocalTableNode(tableData, position);
			}

			// Reconstruct relationship edges from backend data
			for (const relationship of relationships) {
				flowStore.addRelationshipEdge({
					id: relationship.id,
					fromTable: relationship.from_table_id,
					toTable: relationship.to_table_id,
					fromField: relationship.from_field_id,
					toField: relationship.to_field_id,
					type: relationship.relationship_type
				});
			}

			// Apply viewport settings if available
			if ($projectStore.currentProject?.canvas_data) {
				try {
					const canvasData = JSON.parse($projectStore.currentProject.canvas_data);
					if (canvasData.viewport) {
						flowStore.updateViewport(canvasData.viewport);
					}
				} catch (error) {
					console.warn('Failed to apply saved viewport:', error);
				}
			}

			console.log(`DEBUG: Loaded ${tables.length} tables and ${relationships.length} relationships`);
			console.log('DEBUG: Final saved positions used:', savedPositions);
			console.log('DEBUG: Current project canvas_data length:', $projectStore.currentProject?.canvas_data?.length || 0);
		} catch (error) {
			console.error('Failed to load project data:', error);
		}
	}

	onDestroy(() => {
		collaborationStore.disconnect();
	});

	function toggleLeftSidebar() {
		showLeftSidebar = !showLeftSidebar;
	}

	function toggleRightSidebar() {
		showRightSidebar = !showRightSidebar;
	}
</script>

<svelte:head>
	<title>Database Designer - {$projectStore.currentProject?.name || 'Loading...'}</title>
</svelte:head>

<div class="designer-layout h-screen flex flex-col bg-gray-50">
	<!-- Header -->
	<header class="designer-header bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
		<div class="flex items-center space-x-4">
			<a href="/projects/{projectId}" class="text-blue-600 hover:text-blue-800">
				← Back to Project
			</a>
			<div class="h-6 w-px bg-gray-300"></div>
			<h1 class="text-xl font-semibold text-gray-900">
				{$projectStore.currentProject?.name || 'Loading...'}
			</h1>
			<span class="text-sm text-gray-500">Database Designer</span>
		</div>

		<div class="flex items-center space-x-4">
			<CollaborationStatus />
			<PresenceList />
		</div>
	</header>

	<!-- Main Content -->
	<div class="designer-content flex-1 flex overflow-hidden">
		<!-- Left Sidebar -->
		{#if showLeftSidebar}
			<aside class="left-sidebar w-80 bg-white border-r border-gray-200 flex flex-col">
				<div class="p-4 border-b border-gray-200">
					<h2 class="text-lg font-medium text-gray-900">Design Tools</h2>
				</div>
				<div class="flex-1 overflow-y-auto">
					<Toolbar />
					<PropertyPanel />
				</div>
			</aside>
		{/if}

		<!-- Canvas Area -->
		<main class="canvas-area flex-1 relative">
			<div bind:this={canvasContainer} class="w-full h-full">
				<DatabaseCanvas />
			</div>

			<!-- Sidebar Toggle Buttons -->
			<button
				on:click={toggleLeftSidebar}
				class="absolute top-4 left-4 z-10 bg-white shadow-lg rounded-lg p-2 hover:bg-gray-50"
				title={showLeftSidebar ? 'Hide sidebar' : 'Show sidebar'}
			>
				{#if showLeftSidebar}
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					</svg>
				{:else}
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
					</svg>
				{/if}
			</button>

			<button
				on:click={toggleRightSidebar}
				class="absolute top-4 right-4 z-10 bg-white shadow-lg rounded-lg p-2 hover:bg-gray-50"
				title={showRightSidebar ? 'Hide activity' : 'Show activity'}
			>
				{#if showRightSidebar}
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
					</svg>
				{:else}
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
					</svg>
				{/if}
			</button>

			<!-- Live Cursors are now rendered inside DatabaseCanvas/SvelteFlow -->

			<!-- Debug Info -->
			<div class="absolute top-8 left-4 bg-black bg-opacity-75 text-white p-2 rounded text-xs font-mono z-50">
				<div>Connected Users: {$collaborationStore.connectedUsers.length}</div>
				<div>Users with cursors: {$collaborationStore.connectedUsers.filter(u => u.cursor).length}</div>

				<!-- Current User -->
				<div class="mt-1 text-yellow-300">
					{$authStore.user?.username || 'You'} (current):
					{$collaborationStore.currentUserCursor ? `(${$collaborationStore.currentUserCursor.x.toFixed(1)}, ${$collaborationStore.currentUserCursor.y.toFixed(1)})` : 'No cursor'}
				</div>

				<!-- Other Users -->
				{#each $collaborationStore.connectedUsers as user}
					<div class="mt-1">
						{user.username || 'Unknown'}:
						{user.cursor ? `(${user.cursor.x.toFixed(1)}, ${user.cursor.y.toFixed(1)})` : 'No cursor'}
					</div>
				{/each}
			</div>
		</main>

		<!-- Right Sidebar -->
		{#if showRightSidebar}
			<aside class="right-sidebar w-80 bg-white border-l border-gray-200 flex flex-col">
				<div class="p-4 border-b border-gray-200">
					<h2 class="text-lg font-medium text-gray-900">Activity</h2>
				</div>
				<div class="flex-1 overflow-y-auto">
					<ActivityFeed />
				</div>
			</aside>
		{/if}
	</div>
</div>

<style>
	.designer-layout {
		min-height: 100vh;
	}

	.canvas-area {
		background: #f8fafc;
		background-image:
			radial-gradient(circle, #e2e8f0 1px, transparent 1px);
		background-size: 20px 20px;
	}
</style>