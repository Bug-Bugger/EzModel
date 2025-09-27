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

	const projectId = $page.params.id;

	let canvasContainer: HTMLElement;
	let showLeftSidebar = true;
	let showRightSidebar = true;

	onMount(async () => {
		// Load project data
		if (projectId) {
			await projectStore.loadProject(projectId);

			// Initialize WebSocket connection for collaboration
			await collaborationStore.connect(projectId);

			// Load existing canvas data if available
			if ($projectStore.currentProject?.canvas_data) {
				flowStore.loadCanvasData($projectStore.currentProject.canvas_data);
			}
		}
	});

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
			<div class="absolute top-4 left-4 bg-black bg-opacity-75 text-white p-2 rounded text-xs font-mono z-50">
				<div>Connected Users: {$collaborationStore.connectedUsers.length}</div>
				<div>Users with cursors: {$collaborationStore.connectedUsers.filter(u => u.cursor).length}</div>
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