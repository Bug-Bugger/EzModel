<script lang="ts">
	import { designerStore } from '$lib/stores/designer';
	import { collaborationStore } from '$lib/stores/collaboration';
	import Button from '../ui/button.svelte';

	// Tool selection
	function selectTool(tool: 'select' | 'table' | 'relationship') {
		designerStore.selectTool(tool);
	}

	// Export functions
	async function exportSchema(format: 'postgresql' | 'mysql' | 'sqlite' | 'sqlserver') {
		designerStore.startExport(format);

		// TODO: Implement actual export logic
		// This would generate SQL based on the current schema
		console.log(`Exporting schema as ${format}`);

		setTimeout(() => {
			designerStore.finishExport();
		}, 2000);
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

<div class="toolbar p-4 space-y-6">
	<!-- Design Tools -->
	<div class="tool-section">
		<h3 class="text-sm font-medium text-gray-900 mb-3">Design Tools</h3>
		<div class="tool-grid grid grid-cols-1 gap-2">
			<Button
				variant={$designerStore.toolbar.selectedTool === 'select' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('select')}
				class="justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.121 2.122" />
					</svg>
					Select
				{/snippet}
			</Button>

			<Button
				variant={$designerStore.toolbar.selectedTool === 'table' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('table')}
				class="justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M3 14h18m-9-4v8m-7 0V4a1 1 0 011-1h16a1 1 0 011 1v16a1 1 0 01-1 1H5a1 1 0 01-1-1z" />
					</svg>
					Add Table
				{/snippet}
			</Button>

			<Button
				variant={$designerStore.toolbar.selectedTool === 'relationship' ? 'default' : 'outline'}
				size="sm"
				onclick={() => selectTool('relationship')}
				class="justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
					</svg>
					Add Relationship
				{/snippet}
			</Button>
		</div>
	</div>

	<!-- Canvas Settings -->
	<div class="tool-section">
		<h3 class="text-sm font-medium text-gray-900 mb-3">Canvas Settings</h3>
		<div class="space-y-2">
			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.showGrid}
					on:change={toggleGrid}
					class="mr-2 rounded"
				/>
				<span class="text-sm text-gray-700">Show Grid</span>
			</label>

			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.snapToGrid}
					on:change={toggleSnapToGrid}
					class="mr-2 rounded"
				/>
				<span class="text-sm text-gray-700">Snap to Grid</span>
			</label>

			<label class="flex items-center">
				<input
					type="checkbox"
					checked={$designerStore.showMinimap}
					on:change={toggleMinimap}
					class="mr-2 rounded"
				/>
				<span class="text-sm text-gray-700">Show Minimap</span>
			</label>
		</div>
	</div>

	<!-- Export Options -->
	<div class="tool-section">
		<h3 class="text-sm font-medium text-gray-900 mb-3">Export Schema</h3>
		<div class="space-y-2">
			<Button
				variant="outline"
				size="sm"
				onclick={() => exportSchema('postgresql')}
				loading={$designerStore.isExporting && $designerStore.exportFormat === 'postgresql'}
				disabled={$designerStore.isExporting}
				class="w-full justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
					</svg>
					PostgreSQL
				{/snippet}
			</Button>

			<Button
				variant="outline"
				size="sm"
				onclick={() => exportSchema('mysql')}
				loading={$designerStore.isExporting && $designerStore.exportFormat === 'mysql'}
				disabled={$designerStore.isExporting}
				class="w-full justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
					</svg>
					MySQL
				{/snippet}
			</Button>

			<Button
				variant="outline"
				size="sm"
				onclick={() => exportSchema('sqlite')}
				loading={$designerStore.isExporting && $designerStore.exportFormat === 'sqlite'}
				disabled={$designerStore.isExporting}
				class="w-full justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
					</svg>
					SQLite
				{/snippet}
			</Button>

			<Button
				variant="outline"
				size="sm"
				onclick={() => exportSchema('sqlserver')}
				loading={$designerStore.isExporting && $designerStore.exportFormat === 'sqlserver'}
				disabled={$designerStore.isExporting}
				class="w-full justify-start"
			>
				{#snippet children()}
					<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
					</svg>
					SQL Server
				{/snippet}
			</Button>
		</div>
	</div>

	<!-- Quick Actions -->
	<div class="tool-section">
		<h3 class="text-sm font-medium text-gray-900 mb-3">Quick Actions</h3>
		<div class="space-y-2">
			<Button
				variant="outline"
				size="sm"
				class="w-full justify-start"
			>
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
				</svg>
				Fit to View
			</Button>

			<Button
				variant="outline"
				size="sm"
				class="w-full justify-start"
			>
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
				</svg>
				Auto Layout
			</Button>
		</div>
	</div>
</div>

<style>
	.tool-section {
		border-bottom: 1px solid #e5e7eb;
		padding-bottom: 1.5rem;
	}

	.tool-section:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}
</style>