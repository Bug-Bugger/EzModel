<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { projectStore } from '$lib/stores/project';
	import Button from '$lib/components/ui/button.svelte';
	import AddCollaboratorModal from '$lib/components/project/AddCollaboratorModal.svelte';

	const projectId = $page.params.id;

	let showAddCollaboratorModal = $state(false);
	let removingCollaboratorId = $state<string | null>(null);

	onMount(async () => {
		if (projectId) {
			await projectStore.setCurrentProject(projectId);
		}
	});

	async function removeCollaborator(collaboratorId: string) {
		if (removingCollaboratorId) return;

		const confirmed = confirm('Are you sure you want to remove this collaborator?');
		if (!confirmed) return;

		removingCollaboratorId = collaboratorId;
		try {
			await projectStore.removeCollaborator(collaboratorId);
		} catch (error) {
			console.error('Failed to remove collaborator:', error);
			alert('Failed to remove collaborator. Please try again.');
		} finally {
			removingCollaboratorId = null;
		}
	}

	function editSchema() {
		goto(`/projects/${projectId}/edit`);
	}

	function backToProjects() {
		goto('/projects');
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getDatabaseIcon(dbType: string) {
		switch (dbType.toLowerCase()) {
			case 'postgresql':
				return '🐘';
			case 'mysql':
				return '🐬';
			case 'sqlite':
				return '📁';
			case 'sqlserver':
				return '🟦';
			default:
				return '🗄️';
		}
	}
</script>

<svelte:head>
	<title>{$projectStore.currentProject?.name || 'Project'} - EzModel</title>
</svelte:head>

<div class="project-detail max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	{#if $projectStore.isLoading}
		<div class="loading-state flex items-center justify-center py-12">
			<svg class="animate-spin h-8 w-8 text-blue-600" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
				></circle>
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				></path>
			</svg>
			<span class="ml-3 text-gray-600">Loading project...</span>
		</div>
	{:else if $projectStore.currentProject}
		<!-- Navigation -->
		<div class="project-nav mb-6">
			<button
				onclick={backToProjects}
				class="flex items-center text-blue-600 hover:text-blue-800 transition-colors"
			>
				<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M15 19l-7-7 7-7"
					/>
				</svg>
				Back to Projects
			</button>
		</div>

		<!-- Project Header -->
		<div class="project-header bg-white border border-gray-200 rounded-lg shadow-sm p-8 mb-8">
			<div class="flex items-start justify-between">
				<div class="flex-1">
					<div class="flex items-center mb-4">
						<span class="text-3xl mr-3"
							>{getDatabaseIcon($projectStore.currentProject.database_type)}</span
						>
						<div>
							<h1 class="text-3xl font-bold text-gray-900">{$projectStore.currentProject.name}</h1>
							<p class="text-gray-600 mt-1">Database Schema Project</p>
						</div>
					</div>

					{#if $projectStore.currentProject.description}
						<p class="text-gray-700 mb-6 max-w-3xl">{$projectStore.currentProject.description}</p>
					{/if}

					<!-- Project Metadata -->
					<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
						<div class="metadata-item">
							<div class="flex items-center mb-2">
								<svg
									class="w-5 h-5 text-gray-400 mr-2"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4"
									/>
								</svg>
								<span class="text-sm font-medium text-gray-900">Database Type</span>
							</div>
							<span class="text-lg text-blue-600 font-medium"
								>{$projectStore.currentProject.database_type}</span
							>
						</div>

						<div class="metadata-item">
							<div class="flex items-center mb-2">
								<svg
									class="w-5 h-5 text-gray-400 mr-2"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2v2z"
									/>
								</svg>
								<span class="text-sm font-medium text-gray-900">Tables</span>
							</div>
							<span class="text-lg text-green-600 font-medium"
								>{$projectStore.currentProject.tables?.length || 0}</span
							>
						</div>

						<div class="metadata-item">
							<div class="flex items-center mb-2">
								<svg
									class="w-5 h-5 text-gray-400 mr-2"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								<span class="text-sm font-medium text-gray-900">Last Updated</span>
							</div>
							<span class="text-sm text-gray-600"
								>{formatDate($projectStore.currentProject.updated_at)}</span
							>
						</div>
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="flex items-center space-x-3 ml-6">
					<Button variant="outline" size="lg">
						<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z"
							/>
						</svg>
						Share
					</Button>

					<Button size="lg" onclick={editSchema}>
						{#snippet children()}
							<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
								/>
							</svg>
							Edit Schema
						{/snippet}
					</Button>
				</div>
			</div>
		</div>

		<!-- Project Content -->
		<div class="project-content grid grid-cols-1 lg:grid-cols-3 gap-8">
			<!-- Tables Overview -->
			<div class="tables-overview lg:col-span-2">
				<div class="bg-white border border-gray-200 rounded-lg shadow-sm">
					<div class="tables-header p-6 border-b border-gray-200">
						<h2 class="text-xl font-semibold text-gray-900">Database Tables</h2>
						<p class="text-gray-600 mt-1">Overview of your database schema structure</p>
					</div>

					<div class="tables-list p-6">
						{#if $projectStore.currentProject.tables && $projectStore.currentProject.tables.length > 0}
							<div class="space-y-4">
								{#each $projectStore.currentProject.tables as table}
									<div class="table-item bg-gray-50 border border-gray-200 rounded-lg p-4">
										<div class="flex items-center justify-between mb-3">
											<h3 class="text-lg font-medium text-gray-900">{table.name}</h3>
											<span class="text-sm text-gray-500">{table.fields?.length || 0} fields</span>
										</div>

										{#if table.fields && table.fields.length > 0}
											<div class="fields-preview">
												<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
													{#each table.fields.slice(0, 6) as field}
														<div class="field-item bg-white px-3 py-2 rounded border text-sm">
															<span class="font-medium text-gray-900">{field.name}</span>
															<span class="text-gray-500 ml-2">{field.data_type}</span>
															{#if field.is_primary_key}
																<span class="text-yellow-600 ml-1">PK</span>
															{/if}
														</div>
													{/each}
													{#if table.fields.length > 6}
														<div
															class="field-item bg-gray-100 px-3 py-2 rounded border text-sm text-gray-600 flex items-center justify-center"
														>
															+{table.fields.length - 6} more
														</div>
													{/if}
												</div>
											</div>
										{:else}
											<p class="text-gray-500 text-sm">No fields defined</p>
										{/if}
									</div>
								{/each}
							</div>
						{:else}
							<div class="empty-tables text-center py-8">
								<svg
									class="w-16 h-16 mx-auto mb-4 text-gray-300"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="1"
										d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2v2z"
									/>
								</svg>
								<h3 class="text-lg font-medium text-gray-900 mb-2">No tables yet</h3>
								<p class="text-gray-600 mb-4">
									Start building your database schema by adding tables
								</p>
								<Button onclick={editSchema}>
									{#snippet children()}
										Start Designing
									{/snippet}
								</Button>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Project Info Sidebar -->
			<div class="project-sidebar space-y-6">
				<!-- Quick Actions -->
				<div class="bg-white border border-gray-200 rounded-lg shadow-sm p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
					<div class="space-y-3">
						<Button variant="outline" size="sm" class="w-full justify-start" onclick={editSchema}>
							{#snippet children()}
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
									/>
								</svg>
								Edit Schema
							{/snippet}
						</Button>

						<Button variant="outline" size="sm" class="w-full justify-start">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
								/>
							</svg>
							Export SQL
						</Button>

						<Button variant="outline" size="sm" class="w-full justify-start">
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z"
								/>
							</svg>
							Share Project
						</Button>
					</div>
				</div>

				<!-- Project Statistics -->
				<div class="bg-white border border-gray-200 rounded-lg shadow-sm p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Statistics</h3>
					<div class="space-y-4">
						<div class="stat-item flex justify-between">
							<span class="text-gray-600">Tables</span>
							<span class="font-medium text-gray-900"
								>{$projectStore.currentProject.tables?.length || 0}</span
							>
						</div>
						<div class="stat-item flex justify-between">
							<span class="text-gray-600">Total Fields</span>
							<span class="font-medium text-gray-900">
								{$projectStore.currentProject.tables?.reduce(
									(total: number, table: any) => total + (table.fields?.length || 0),
									0
								) || 0}
							</span>
						</div>
						<div class="stat-item flex justify-between">
							<span class="text-gray-600">Relationships</span>
							<span class="font-medium text-gray-900"
								>{$projectStore.currentProject.relationships?.length || 0}</span
							>
						</div>
					</div>
				</div>

				<!-- Collaborators -->
				<div class="bg-white border border-gray-200 rounded-lg shadow-sm p-6">
					<h3 class="text-lg font-semibold text-gray-900 mb-4">Collaborators</h3>
					{#if $projectStore.currentProject.collaborators && $projectStore.currentProject.collaborators.length > 0}
						<div class="space-y-3">
							{#each $projectStore.currentProject.collaborators as collaborator}
								<div class="collaborator-item flex items-center space-x-3">
									<div
										class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium"
									>
										{(collaborator.username || 'U').charAt(0).toUpperCase()}
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-gray-900 truncate">
											{collaborator.username || 'Unknown User'}
										</p>
										<p class="text-xs text-gray-500 truncate">{collaborator.email || 'No email'}</p>
									</div>
									<button
										class="p-1 text-gray-400 hover:text-red-500 transition-colors"
										onclick={() => removeCollaborator(collaborator.id)}
										disabled={removingCollaboratorId === collaborator.id}
										title="Remove collaborator"
									>
										{#if removingCollaboratorId === collaborator.id}
											<svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
												<circle
													class="opacity-25"
													cx="12"
													cy="12"
													r="10"
													stroke="currentColor"
													stroke-width="4"
												></circle>
												<path
													class="opacity-75"
													fill="currentColor"
													d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
												></path>
											</svg>
										{:else}
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path
													stroke-linecap="round"
													stroke-linejoin="round"
													stroke-width="2"
													d="M6 18L18 6M6 6l12 12"
												/>
											</svg>
										{/if}
									</button>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-gray-500 text-sm">No collaborators yet</p>
					{/if}

					<Button
						variant="outline"
						size="sm"
						class="w-full mt-4"
						onclick={() => (showAddCollaboratorModal = true)}
					>
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 6v6m0 0v6m0-6h6m-6 0H6"
							/>
						</svg>
						Add Collaborator
					</Button>
				</div>
			</div>
		</div>
	{:else}
		<div class="error-state text-center py-12">
			<svg
				class="w-16 h-16 mx-auto mb-4 text-red-300"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1"
					d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
			<h3 class="text-xl font-medium text-gray-900 mb-2">Project not found</h3>
			<p class="text-gray-600 mb-6">
				The project you're looking for doesn't exist or you don't have access to it.
			</p>
			<Button onclick={backToProjects}>
				{#snippet children()}
					Back to Projects
				{/snippet}
			</Button>
		</div>
	{/if}
</div>

<!-- Add Collaborator Modal -->
<AddCollaboratorModal
	bind:open={showAddCollaboratorModal}
	onOpenChange={(open) => (showAddCollaboratorModal = open)}
/>

<style>
	.metadata-item {
		transition: transform 0.2s;
	}

	.metadata-item:hover {
		transform: translateY(-1px);
	}

	.table-item:hover {
		background-color: #f9fafb;
		border-color: #d1d5db;
	}
</style>
