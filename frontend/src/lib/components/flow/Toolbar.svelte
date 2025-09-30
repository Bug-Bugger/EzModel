<script lang="ts">
	import { designerStore } from '$lib/stores/designer';
	import Button from '../ui/button.svelte';

	// Tool selection
	function selectTool(tool: 'select' | 'table' | 'relationship') {
		designerStore.selectTool(tool);
	}

	// Canvas controls
	function toggleGrid() {
		designerStore.toggleGrid();
	}

	function toggleSnapToGrid() {
		designerStore.toggleSnapToGrid();
	}

	function toggleMinimap() {
		designerStore.toggleMinimap();
	}
</script>

<div class="toolbar p-3 space-y-4">
	<!-- Design Tools -->
	<div class="tool-section">
		<div class="tool-grid grid grid-cols-3 gap-1">
			<Button
				variant={$designerStore.toolbar.selectedTool === 'select' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('select')}
				class="p-2 h-8"
			>
				{#snippet children()}
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.121 2.122"
						/>
					</svg>
				{/snippet}
			</Button>

			<Button
				variant={$designerStore.toolbar.selectedTool === 'table' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('table')}
				class="p-2 h-8"
			>
				{#snippet children()}
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M3 10h18M3 14h18m-9-4v8m-7 0V4a1 1 0 011-1h16a1 1 0 011 1v16a1 1 0 01-1 1H5a1 1 0 01-1-1z"
						/>
					</svg>
				{/snippet}
			</Button>

			<Button
				variant={$designerStore.toolbar.selectedTool === 'relationship' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('relationship')}
				class="p-2 h-8"
			>
				{#snippet children()}
					<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
						/>
					</svg>
				{/snippet}
			</Button>
		</div>
	</div>

	<!-- Canvas Settings -->
	<div class="tool-section">
		<h3 class="text-xs font-medium text-gray-900 mb-2">Canvas Settings</h3>
		<div class="space-y-1">
			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.showGrid}
					on:change={toggleGrid}
					class="mr-2 rounded w-3 h-3"
				/>
				<span class="text-xs text-gray-700">Show Grid</span>
			</label>

			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.snapToGrid}
					on:change={toggleSnapToGrid}
					class="mr-2 rounded w-3 h-3"
				/>
				<span class="text-xs text-gray-700">Snap to Grid</span>
			</label>

			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.showMinimap}
					on:change={toggleMinimap}
					class="mr-2 rounded w-3 h-3"
				/>
				<span class="text-xs text-gray-700">Show Minimap</span>
			</label>
		</div>
	</div>
</div>

<style>
	.tool-section {
		border-bottom: 1px solid #e5e7eb;
		padding-bottom: 1rem;
	}

	.tool-section:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}
</style>
