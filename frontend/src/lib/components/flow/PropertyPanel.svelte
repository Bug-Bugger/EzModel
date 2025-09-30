<script lang="ts">
	import { designerStore } from '$lib/stores/designer';
	import { flowStore } from '$lib/stores/flow';
	import { collaborationStore } from '$lib/stores/collaboration';
	import { projectStore } from '$lib/stores/project';
	import { projectService } from '$lib/services/project';
	import Button from '../ui/button.svelte';
	import Input from '../ui/input.svelte';
	import Select from '../ui/select.svelte';
	import type { CreateFieldRequest, UpdateFieldRequest } from '$lib/types/models';

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
		{ value: 'one_to_one', label: 'One to One (1:1)' },
		{ value: 'one_to_many', label: 'One to Many (1:N)' },
		{ value: 'many_to_many', label: 'Many to Many (N:M)' }
	];

	// Reactive property panel state
	$: panel = $designerStore.propertyPanel;
	$: selectedNode = $flowStore.selectedNode;
	$: selectedEdge = $flowStore.selectedEdge;

	// Explicit reactive tracking for fields to ensure UI updates
	$: tableFields = selectedNode?.data?.fields || [];

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
	$: if (selectedNode && selectedNode.data) {
		tableName = selectedNode.data.name || '';
	} else {
		tableName = '';
	}

	// Helper function to prepare field data for backend API
	function prepareFieldForBackend(field: any): CreateFieldRequest {
		return {
			name: field.name,
			data_type: field.data_type,
			is_primary_key: field.is_primary_key,
			is_nullable: field.is_nullable,
			default_value: field.default_value || '',
			position: (selectedNode && selectedNode.data) ? selectedNode.data.fields.length : 0
		};
	}

	// Update table name
	function updateTableName() {
		if (selectedNode && selectedNode.data && tableName !== selectedNode.data.name) {
			flowStore.updateTableNode(selectedNode.id, { name: tableName });
			collaborationStore.sendSchemaEvent('table_update', {
				id: selectedNode.id,
				name: tableName
			});
		}
	}

	// Add new field to table
	async function addField() {
		if (!selectedNode || !selectedNode.data || !fieldName.trim() || !$projectStore.currentProject) {
			return;
		}

		try {
			// Create field data using backend structure
			const fieldData = {
				name: fieldName.trim(),
				data_type: fieldType,
				is_primary_key: isPrimary,
				is_nullable: !isRequired, // Convert frontend "required" to backend "nullable" (inverse)
				default_value: defaultValue || '',
				position: tableFields.length
			};

			// Create via API using backend format
			const createdField = await projectService.createField(
				$projectStore.currentProject.id,
				selectedNode.id,
				fieldData
			);

			// Update local store with the returned field (already in correct format)
			const updatedFields = [...selectedNode.data.fields, createdField];
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data to persist field changes
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);

			// Reset form
			fieldName = '';
			fieldType = 'STRING';
			isPrimary = false;
			isForeign = false;
			isRequired = false;
			isUnique = false;
			defaultValue = '';
		} catch (error) {
			console.error('Failed to create field:', error);
			// TODO: Show error message to user
		}
	}

	// Remove field from table
	async function removeField(fieldId: string) {
		if (!selectedNode || !selectedNode.data || !$projectStore.currentProject) {
			return;
		}

		try {
			const field = selectedNode.data.fields.find(f => f.field_id === fieldId);
			if (!field) {
				return;
			}

			// Delete field via API
			await projectService.deleteField(
				$projectStore.currentProject.id,
				selectedNode.id,
				fieldId
			);

			// Update local store
			const updatedFields = selectedNode.data.fields.filter(f => f.field_id !== fieldId);
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data to persist field deletion
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);
		} catch (error) {
			console.error('Failed to delete field:', error);
			// TODO: Show error message to user
		}
	}

	// Update existing field
	async function updateField(fieldId: string, updates: any) {
		if (!selectedNode || !selectedNode.data || !$projectStore.currentProject) {
			return;
		}

		try {
			const currentField = selectedNode.data.fields.find(f => f.field_id === fieldId);
			if (!currentField) {
				return;
			}

			// Prepare updates in backend format (updates should already be in backend format)
			const backendUpdates: UpdateFieldRequest = {};
			if (updates.name !== undefined) backendUpdates.name = updates.name;
			if (updates.data_type !== undefined) backendUpdates.data_type = updates.data_type;
			if (updates.is_primary_key !== undefined) backendUpdates.is_primary_key = updates.is_primary_key;
			if (updates.is_nullable !== undefined) backendUpdates.is_nullable = updates.is_nullable;
			if (updates.default_value !== undefined) backendUpdates.default_value = updates.default_value;

			// Update field via API
			const updatedField = await projectService.updateField(
				$projectStore.currentProject.id,
				selectedNode.id,
				fieldId,
				backendUpdates
			);

			// Update local store with the returned field (already in correct format)
			const updatedFields = selectedNode.data.fields.map(field =>
				field.field_id === fieldId ? updatedField : field
			);
			flowStore.updateTableNode(selectedNode.id, { fields: updatedFields });

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data to persist field updates
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);
		} catch (error) {
			console.error('Failed to update field:', error);
			// TODO: Show error message to user
		}
	}

	// Update relationship type
	async function updateRelationshipType(relationshipType: string) {
		if (!selectedEdge || !selectedEdge.data || !$projectStore.currentProject) {
			return;
		}

		try {
			// Update relationship via API
			await flowStore.updateRelationshipEdge(
				$projectStore.currentProject.id,
				selectedEdge.id,
				{ relation_type: relationshipType as 'one_to_one' | 'one_to_many' | 'many_to_many' }
			);

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);
		} catch (error) {
			console.error('Failed to update relationship:', error);
			// TODO: Show error message to user
		}
	}

	// Delete relationship
	async function deleteRelationship() {
		if (!selectedEdge || !selectedEdge.data || !$projectStore.currentProject) {
			return;
		}

		try {
			// Delete relationship via API
			await flowStore.removeRelationshipEdge(
				$projectStore.currentProject.id,
				selectedEdge.id
			);

			// WebSocket broadcasting is now handled by the backend after successful API call

			// Auto-save canvas data
			const canvasData = flowStore.getCurrentCanvasData();
			projectStore.autoSaveCanvasData(canvasData);
		} catch (error) {
			console.error('Failed to delete relationship:', error);
			// TODO: Show error message to user
		}
	}
</script>

<div class="property-panel p-4">
	{#if panel.isOpen && panel.type === 'table' && selectedNode && selectedNode.data}
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
					{#each tableFields as field}
						<div class="field-item bg-gray-50 p-3 rounded border">
							<div class="flex items-center justify-between mb-2">
								<span class="font-medium text-gray-900">{field.name}</span>
								<button
									on:click={() => removeField(field.field_id)}
									class="text-red-600 hover:text-red-800 text-sm"
								>
									Remove
								</button>
							</div>
							<div class="flex items-center space-x-2 text-xs text-gray-600">
								<span class="bg-white px-2 py-1 rounded">{field.data_type}</span>
								{#if field.is_primary_key}
									<span class="bg-yellow-100 text-yellow-800 px-2 py-1 rounded">PK</span>
								{/if}
								{#if !field.is_nullable}
									<span class="bg-red-100 text-red-800 px-2 py-1 rounded">NOT NULL</span>
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
						bind:value={selectedEdge.data.relation_type}
						options={relationshipTypeOptions}
						onchange={(value) => updateRelationshipType(value)}
						class="w-full"
					/>
				</div>

				<div>
					<label for="from-table" class="block text-sm font-medium text-gray-700 mb-2">From Table</label>
					<Input id="from-table" value={selectedEdge.data.source_table_id} disabled class="w-full" />
				</div>

				<div>
					<label for="to-table" class="block text-sm font-medium text-gray-700 mb-2">To Table</label>
					<Input id="to-table" value={selectedEdge.data.target_table_id} disabled class="w-full" />
				</div>

				<div>
					<label for="from-field" class="block text-sm font-medium text-gray-700 mb-2">From Field</label>
					<Input id="from-field" value={selectedEdge.data.source_field_id} disabled class="w-full" />
				</div>

				<div>
					<label for="to-field" class="block text-sm font-medium text-gray-700 mb-2">To Field</label>
					<Input id="to-field" value={selectedEdge.data.target_field_id} disabled class="w-full" />
				</div>

				<!-- Delete Relationship Button -->
				<div class="border-t pt-4">
					<Button
						onclick={deleteRelationship}
						variant="destructive"
						size="sm"
						class="w-full"
					>
						{#snippet children()}
							Delete Relationship
						{/snippet}
					</Button>
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