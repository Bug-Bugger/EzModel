<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Handle, Position } from '@xyflow/svelte';
	import type { TableNode } from '$lib/stores/flow';
	import { designerStore } from '$lib/stores/designer';

	export let data: TableNode['data'];
	export let selected: boolean = false;

	const dispatch = createEventDispatcher();

	// Show handles only when relationship tool is selected (for visual feedback)
	$: showHandles = $designerStore.toolbar.selectedTool === 'relationship';

	// Icons for different field types and constraints
	function getFieldIcon(field: any) {
		if (field.is_primary) return '🔑';
		if (field.is_foreign) return '🔗';
		return '📄';
	}

	function getFieldTypeColor(type: string) {
		const typeColors: { [key: string]: string } = {
			UUID: 'text-purple-600',
			STRING: 'text-green-600',
			INTEGER: 'text-blue-600',
			BOOLEAN: 'text-yellow-600',
			TIMESTAMP: 'text-red-600',
			TEXT: 'text-gray-600',
			DECIMAL: 'text-indigo-600'
		};
		return typeColors[type] || 'text-gray-500';
	}

	// Handle add field button click
	function handleAddField(event: MouseEvent) {
		event.stopPropagation(); // Prevent node selection
		dispatch('addField', { tableId: data.table_id, tableName: data.name });
	}
</script>

<div
	class="table-node bg-white border-2 border-gray-200 rounded-lg shadow-lg min-w-[200px] max-w-[300px]"
	class:selected
>
	<!-- Table Header -->
	<div class="table-header bg-blue-50 px-4 py-3 border-b border-gray-200 rounded-t-lg">
		<div class="flex items-center justify-between">
			<h3 class="font-semibold text-gray-900 truncate">{data.name}</h3>
			<div class="flex items-center space-x-1">
				<!-- Database type indicator -->
				<span class="text-xs text-gray-500 bg-gray-100 px-2 py-1 rounded"> TABLE </span>
			</div>
		</div>
	</div>

	<!-- Fields List -->
	<div class="fields-list max-h-60 overflow-y-auto">
		{#each data.fields as field, index}
			<div
				class="field-row flex items-center px-4 py-2 border-b border-gray-100 hover:bg-gray-50 last:border-b-0 relative"
				class:bg-blue-50={field.is_primary_key}
			>
				<!-- Connection handles for relationships - always present but only visible when relationship tool is selected -->
				<Handle
					type="source"
					position={Position.Right}
					id="{data.table_id}-{field.field_id}-source"
					class="field-handle-source"
					style="position: absolute; right: -8px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; background: #3b82f6; border: 2px solid white; border-radius: 50%; z-index: 10; cursor: pointer; {showHandles
						? ''
						: 'opacity: 0; pointer-events: none;'}"
				/>
				<Handle
					type="target"
					position={Position.Left}
					id="{data.table_id}-{field.field_id}-target"
					class="field-handle-target"
					style="position: absolute; left: -8px; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; background: #10b981; border: 2px solid white; border-radius: 50%; z-index: 10; cursor: pointer; {showHandles
						? ''
						: 'opacity: 0; pointer-events: none;'}"
				/>

				<!-- Field Icon -->
				<span class="field-icon text-sm mr-2">{getFieldIcon(field)}</span>

				<!-- Field Details -->
				<div class="field-details flex-1 min-w-0">
					<div class="flex items-center justify-between">
						<span class="field-name font-medium text-gray-900 truncate">
							{field.name}
						</span>
						<span class="field-type text-xs {getFieldTypeColor(field.data_type)} font-mono">
							{field.data_type}
						</span>
					</div>

					<!-- Field Constraints -->
					{#if field.is_primary_key || !field.is_nullable}
						<div class="field-constraints flex space-x-1 mt-1">
							{#if field.is_primary_key}
								<span class="constraint-badge bg-yellow-100 text-yellow-800 text-xs px-1 rounded"
									>PK</span
								>
							{/if}
							{#if !field.is_nullable}
								<span class="constraint-badge bg-red-100 text-red-800 text-xs px-1 rounded"
									>NOT NULL</span
								>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>

	<!-- Add Field Button -->
	<div class="table-footer p-2 border-t border-gray-200 bg-gray-50 rounded-b-lg">
		<button
			class="add-field-btn w-full text-sm text-gray-600 hover:text-gray-800 py-1 px-2 rounded hover:bg-gray-100 transition-colors"
			on:click={handleAddField}
		>
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

	/* Field handle styling - enhanced for better clickability */
	:global(.field-handle-source) {
		width: 16px;
		height: 16px;
		background: #3b82f6;
		border: 2px solid white;
		border-radius: 50%;
		right: -8px;
		top: 50%;
		transform: translateY(-50%);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		opacity: 1;
		z-index: 10;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	:global(.field-handle-target) {
		width: 16px;
		height: 16px;
		background: #10b981;
		border: 2px solid white;
		border-radius: 50%;
		left: -8px;
		top: 50%;
		transform: translateY(-50%);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		opacity: 1;
		z-index: 10;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	:global(.field-handle-source:hover),
	:global(.field-handle-target:hover) {
		transform: translateY(-50%) scale(1.1);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
	}

	:global(.field-handle-source:active),
	:global(.field-handle-target:active) {
		transform: translateY(-50%) scale(0.95);
	}
</style>
