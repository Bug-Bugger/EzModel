<script lang="ts">
	import { Handle, Position } from '@xyflow/svelte';
	import type { TableNode } from '$lib/stores/flow';

	export let data: TableNode['data'];
	export let selected: boolean = false;

	// Icons for different field types and constraints
	function getFieldIcon(field: any) {
		if (field.isPrimary) return '🔑';
		if (field.isForeign) return '🔗';
		return '📄';
	}

	function getFieldTypeColor(type: string) {
		const typeColors: { [key: string]: string } = {
			'UUID': 'text-purple-600',
			'STRING': 'text-green-600',
			'INTEGER': 'text-blue-600',
			'BOOLEAN': 'text-yellow-600',
			'TIMESTAMP': 'text-red-600',
			'TEXT': 'text-gray-600',
			'DECIMAL': 'text-indigo-600'
		};
		return typeColors[type] || 'text-gray-500';
	}
</script>

<div class="table-node bg-white border-2 border-gray-200 rounded-lg shadow-lg min-w-[200px] max-w-[300px]" class:selected>
	<!-- Table Header -->
	<div class="table-header bg-blue-50 px-4 py-3 border-b border-gray-200 rounded-t-lg">
		<div class="flex items-center justify-between">
			<h3 class="font-semibold text-gray-900 truncate">{data.name}</h3>
			<div class="flex items-center space-x-1">
				<!-- Database type indicator -->
				<span class="text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded">
					TABLE
				</span>
			</div>
		</div>
	</div>

	<!-- Fields List -->
	<div class="fields-list max-h-60 overflow-y-auto">
		{#each data.fields as field, index}
			<div
				class="field-row flex items-center px-4 py-2 border-b border-gray-100 hover:bg-gray-50 last:border-b-0"
				class:bg-blue-50={field.isPrimary}
			>
				<!-- Connection handles for relationships -->
				<Handle
					type="source"
					position={Position.Right}
					id="{data.id}-{field.id}-source"
					class="w-2 h-2 bg-blue-500 border-2 border-white"
					style="top: {(index + 1) * 40 + 20}px"
				/>
				<Handle
					type="target"
					position={Position.Left}
					id="{data.id}-{field.id}-target"
					class="w-2 h-2 bg-green-500 border-2 border-white"
					style="top: {(index + 1) * 40 + 20}px"
				/>

				<!-- Field Icon -->
				<span class="field-icon text-sm mr-2">{getFieldIcon(field)}</span>

				<!-- Field Details -->
				<div class="field-details flex-1 min-w-0">
					<div class="flex items-center justify-between">
						<span class="field-name font-medium text-gray-900 truncate">
							{field.name}
						</span>
						<span class="field-type text-xs {getFieldTypeColor(field.type)} font-mono">
							{field.type}
						</span>
					</div>

					<!-- Field Constraints -->
					{#if field.isPrimary || field.isForeign || field.isRequired || field.isUnique}
						<div class="field-constraints flex space-x-1 mt-1">
							{#if field.isPrimary}
								<span class="constraint-badge bg-yellow-100 text-yellow-800 text-xs px-1 rounded">PK</span>
							{/if}
							{#if field.isForeign}
								<span class="constraint-badge bg-green-100 text-green-800 text-xs px-1 rounded">FK</span>
							{/if}
							{#if field.isRequired}
								<span class="constraint-badge bg-red-100 text-red-800 text-xs px-1 rounded">NOT NULL</span>
							{/if}
							{#if field.isUnique}
								<span class="constraint-badge bg-blue-100 text-blue-800 text-xs px-1 rounded">UNIQUE</span>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>

	<!-- Add Field Button -->
	<div class="table-footer p-2 border-t border-gray-200 bg-gray-50 rounded-b-lg">
		<button class="add-field-btn w-full text-sm text-gray-600 hover:text-gray-800 py-1 px-2 rounded hover:bg-gray-100 transition-colors">
			+ Add Field
		</button>
	</div>
</div>

<style>
	.table-node.selected {
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.field-row {
		position: relative;
	}

	.constraint-badge {
		font-size: 10px;
		line-height: 1.2;
	}

	:global(.svelte-flow__handle) {
		opacity: 0;
		transition: opacity 0.2s;
	}

	.table-node:hover :global(.svelte-flow__handle) {
		opacity: 1;
	}

	.table-node.selected :global(.svelte-flow__handle) {
		opacity: 1;
	}
</style>