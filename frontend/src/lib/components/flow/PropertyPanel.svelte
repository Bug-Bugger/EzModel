<script lang="ts">
	import { designerStore } from '$lib/stores/designer';
	import { flowStore } from '$lib/stores/flow';
	import { collaborationStore } from '$lib/stores/collaboration';
	import Button from '../ui/button.svelte';
	import Input from '../ui/input.svelte';
	import Select from '../ui/select.svelte';

	// Field types available
	const fieldTypeOptions = [
		{ value: 'UUID', label: 'UUID' },
		{ value: 'STRING', label: 'STRING' },
		{ value: 'INTEGER', label: 'INTEGER' },
		{ value: 'BOOLEAN', label: 'BOOLEAN' },
		{ value: 'TIMESTAMP', label: 'TIMESTAMP' },
		{ value: 'TEXT', label: 'TEXT' },
		{ value: 'DECIMAL', label: 'DECIMAL' },
		{ value: 'FLOAT', label: 'FLOAT' },
		{ value: 'DATE', label: 'DATE' },
		{ value: 'JSON', label: 'JSON' }
	];

	// Relationship type options
	const relationshipTypeOptions = [
		{ value: 'one-to-one', label: 'One to One (1:1)' },
		{ value: 'one-to-many', label: 'One to Many (1:N)' },
		{ value: 'many-to-many', label: 'Many to Many (N:M)' }
	];

	// Reactive property panel state
	$: panel = $designerStore.propertyPanel;
	$: selectedNode = $flowStore.selectedNode;
	$: selectedEdge = $flowStore.selectedEdge;

	// Form data
	let tableName = '';
	let fieldName = '';
	let fieldType = 'STRING';
	let isPrimary = false;
	let isForeign = false;
	let isRequired = false;
	let isUnique = false;
	let defaultValue = '';

	// Update form when selection changes
	$: if (selectedNode) {
		tableName = selectedNode.data.name;
	}

	// Update table name
	function updateTableName() {
		if (selectedNode && tableName !== selectedNode.data.name) {
			flowStore.updateTableNode(selectedNode.id, { name: tableName });
			collaborationStore.sendSchemaEvent('table_update', {
				id: selectedNode.id,
				name: tableName
			});
		}
	}

	// Add new field to table
	function addField() {
		if (selectedNode && fieldName.trim()) {
			const newField = {
				id: crypto.randomUUID(),
				name: fieldName.trim(),
				type: fieldType,
				isPrimary,
				isForeign,
				isRequired,
				isUnique,
				defaultValue: defaultValue || undefined
			};

			const updatedFields = [...selectedNode.data.fields, newField];
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			collaborationStore.sendSchemaEvent('field_create', {
				...newField,
				table_id: selectedNode.id,
				table_name: selectedNode.data.name
			});

			// Reset form
			fieldName = '';
			fieldType = 'STRING';
			isPrimary = false;
			isForeign = false;
			isRequired = false;
			isUnique = false;
			defaultValue = '';
		}
	}

	// Remove field from table
	function removeField(fieldId: string) {
		if (selectedNode) {
			const field = selectedNode.data.fields.find(f => f.id === fieldId);
			const updatedFields = selectedNode.data.fields.filter(f => f.id !== fieldId);
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			if (field) {
				collaborationStore.sendSchemaEvent('field_delete', {
					id: fieldId,
					name: field.name,
					table_id: selectedNode.id,
					table_name: selectedNode.data.name
				});
			}
		}
	}

	// Update existing field
	function updateField(fieldId: string, updates: any) {
		if (selectedNode) {
			const updatedFields = selectedNode.data.fields.map(field =>
				field.id === fieldId ? { ...field, ...updates } : field
			);
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			collaborationStore.sendSchemaEvent('field_update', {
				id: fieldId,
				...updates,
				table_id: selectedNode.id,
				table_name: selectedNode.data.name
			});
		}
	}
</script>

<div class="property-panel p-4">
	{#if panel.isOpen && panel.type === 'table' && selectedNode}
		<!-- Table Properties -->
		<div class="property-section">
			<h3 class="text-lg font-medium text-gray-900 mb-4">Table Properties</h3>

			<!-- Table Name -->
			<div class="mb-4">
				<label for="table-name" class="block text-sm font-medium text-gray-700 mb-2">Table Name</label>
				<Input
					id="table-name"
					bind:value={tableName}
					placeholder="Enter table name"
					class="w-full"
				/>
			</div>

			<!-- Existing Fields -->
			<div class="mb-6">
				<h4 class="text-sm font-medium text-gray-700 mb-3">Fields</h4>
				<div class="space-y-2 max-h-40 overflow-y-auto">
					{#each selectedNode.data.fields as field}
						<div class="field-item bg-gray-50 p-3 rounded border">
							<div class="flex items-center justify-between mb-2">
								<span class="font-medium text-gray-900">{field.name}</span>
								<button
									on:click={() => removeField(field.id)}
									class="text-red-600 hover:text-red-800 text-sm"
								>
									Remove
								</button>
							</div>
							<div class="flex items-center space-x-2 text-xs text-gray-600">
								<span class="bg-white px-2 py-1 rounded">{field.type}</span>
								{#if field.isPrimary}
									<span class="bg-yellow-100 text-yellow-800 px-2 py-1 rounded">PK</span>
								{/if}
								{#if field.isForeign}
									<span class="bg-green-100 text-green-800 px-2 py-1 rounded">FK</span>
								{/if}
								{#if field.isRequired}
									<span class="bg-red-100 text-red-800 px-2 py-1 rounded">Required</span>
								{/if}
								{#if field.isUnique}
									<span class="bg-blue-100 text-blue-800 px-2 py-1 rounded">Unique</span>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Add New Field -->
			<div class="add-field-section border-t pt-4">
				<h4 class="text-sm font-medium text-gray-700 mb-3">Add New Field</h4>

				<div class="space-y-3">
					<!-- Field Name -->
					<div>
						<label for="field-name" class="block text-xs font-medium text-gray-600 mb-1">Field Name</label>
						<Input
							id="field-name"
							bind:value={fieldName}
							placeholder="Enter field name"
							class="w-full text-sm"
						/>
					</div>

					<!-- Field Type -->
					<div>
						<label for="field-type" class="block text-xs font-medium text-gray-600 mb-1">Type</label>
						<Select
							bind:value={fieldType}
							options={fieldTypeOptions}
							class="w-full text-sm"
						/>
					</div>

					<!-- Field Constraints -->
					<div class="grid grid-cols-2 gap-2">
						<label class="flex items-center text-xs">
							<input type="checkbox" bind:checked={isPrimary} class="mr-1" />
							Primary Key
						</label>
						<label class="flex items-center text-xs">
							<input type="checkbox" bind:checked={isForeign} class="mr-1" />
							Foreign Key
						</label>
						<label class="flex items-center text-xs">
							<input type="checkbox" bind:checked={isRequired} class="mr-1" />
							Required
						</label>
						<label class="flex items-center text-xs">
							<input type="checkbox" bind:checked={isUnique} class="mr-1" />
							Unique
						</label>
					</div>

					<!-- Default Value -->
					<div>
						<label for="field-default" class="block text-xs font-medium text-gray-600 mb-1">Default Value</label>
						<Input
							id="field-default"
							bind:value={defaultValue}
							placeholder="Enter default value"
							class="w-full text-sm"
						/>
					</div>

					<!-- Add Button -->
					<Button
						onclick={addField}
						disabled={!fieldName.trim()}
						size="sm"
						class="w-full"
					>
						{#snippet children()}
							Add Field
						{/snippet}
					</Button>
				</div>
			</div>
		</div>

	{:else if panel.isOpen && panel.type === 'relationship' && selectedEdge}
		<!-- Relationship Properties -->
		<div class="property-section">
			<h3 class="text-lg font-medium text-gray-900 mb-4">Relationship Properties</h3>

			<div class="space-y-4">
				<div>
					<label for="relationship-type" class="block text-sm font-medium text-gray-700 mb-2">Relationship Type</label>
					<Select
						options={relationshipTypeOptions}
						class="w-full"
					/>
				</div>

				<div>
					<label for="from-table" class="block text-sm font-medium text-gray-700 mb-2">From Table</label>
					<Input id="from-table" value={selectedEdge.data.fromTable} disabled class="w-full" />
				</div>

				<div>
					<label for="to-table" class="block text-sm font-medium text-gray-700 mb-2">To Table</label>
					<Input id="to-table" value={selectedEdge.data.toTable} disabled class="w-full" />
				</div>

				<div>
					<label for="from-field" class="block text-sm font-medium text-gray-700 mb-2">From Field</label>
					<Input id="from-field" value={selectedEdge.data.fromField} class="w-full" />
				</div>

				<div>
					<label for="to-field" class="block text-sm font-medium text-gray-700 mb-2">To Field</label>
					<Input id="to-field" value={selectedEdge.data.toField} class="w-full" />
				</div>
			</div>
		</div>

	{:else}
		<!-- No Selection -->
		<div class="property-section">
			<div class="text-center py-8 text-gray-500">
				<svg class="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.121 2.122" />
				</svg>
				<p class="text-sm">Select a table or relationship to edit properties</p>
			</div>
		</div>
	{/if}
</div>

<style>
	.property-section {
		border-bottom: 1px solid #e5e7eb;
		padding-bottom: 1.5rem;
	}

	.property-section:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	.field-item {
		transition: background-color 0.2s;
	}

	.field-item:hover {
		background-color: #f3f4f6;
	}
</style>