<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { projectStore } from '$lib/stores/project';
	import { projectService } from '$lib/services/project';
	import { uiStore } from '$lib/stores/ui';
	import Button from '$lib/components/ui/button.svelte';
	import AlertDialog from '$lib/components/ui/alert-dialog.svelte';
	import Dialog from '$lib/components/ui/dialog.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import Select from '$lib/components/ui/select.svelte';
	import type { CreateProjectRequest } from '$lib/types/models';

	// Create project dialog state
	let showCreateDialog = false;
	let projectName = '';
	let projectDescription = '';
	let databaseType = 'postgresql';
	let isCreating = false;

	// Delete confirmation dialog state
	let showDeleteDialog = false;
	let projectToDelete: { id: string; name: string } | null = null;

	// Database type options
	const databaseOptions = [
		{ value: 'postgresql', label: 'PostgreSQL' },
		{ value: 'mysql', label: 'MySQL' },
		{ value: 'sqlite', label: 'SQLite' },
		{ value: 'sqlserver', label: 'SQL Server' }
	];

	onMount(() => {
		projectStore.loadProjects();
	});

	function createNewProject() {
		openCreateDialog();
	}

	function openCreateDialog() {
		showCreateDialog = true;
		projectName = '';
		projectDescription = '';
		databaseType = 'postgresql';
		isCreating = false;
	}

	function closeCreateDialog() {
		showCreateDialog = false;
		projectName = '';
		projectDescription = '';
		databaseType = 'postgresql';
		isCreating = false;
	}

	async function createProject() {
		if (!projectName.trim()) {
			uiStore.error('Project name is required');
			return;
		}

		isCreating = true;
		try {
			const projectData: CreateProjectRequest = {
				name: projectName.trim(),
				description: projectDescription.trim() || undefined,
				database_type: databaseType as any
			};

			const newProject = await projectService.createProject(projectData);
			projectStore.addProject(newProject);
			uiStore.success('Project created successfully!');
			closeCreateDialog();
		} catch (error: any) {
			uiStore.error('Failed to create project', error.message);
		} finally {
			isCreating = false;
		}
	}

	function openProject(projectId: string) {
		goto(`/projects/${projectId}`);
	}

	function editProject(projectId: string) {
		goto(`/projects/${projectId}/edit`);
	}

	function openDeleteDialog(project: { id: string; name: string }) {
		projectToDelete = project;
		showDeleteDialog = true;
	}

	async function confirmDeleteProject() {
		if (!projectToDelete) return;

		try {
			await projectService.deleteProject(projectToDelete.id);
			projectStore.removeProject(projectToDelete.id);
			uiStore.success('Project deleted successfully');
			showDeleteDialog = false;
			projectToDelete = null;
		} catch (error: any) {
			uiStore.error('Failed to delete project', error.message);
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString();
	}
</script>

<svelte:head>
	<title>Projects - EzModel</title>
</svelte:head>

<div class="projects-page max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<!-- Header -->
	<div class="projects-header flex items-center justify-between mb-8">
		<div>
			<h1 class="text-3xl font-bold text-gray-900">My Projects</h1>
			<p class="text-gray-600 mt-2">Design and manage your database schemas</p>
		</div>

		<Button onclick={createNewProject}>
			{#snippet children()}
				<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
				</svg>
				New Project
			{/snippet}
		</Button>
	</div>

	<!-- Projects Grid -->
	{#if $projectStore.isLoading}
		<div class="loading-state flex items-center justify-center py-12">
			<svg class="animate-spin h-8 w-8 text-blue-600" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
			</svg>
			<span class="ml-3 text-gray-600">Loading projects...</span>
		</div>
	{:else if $projectStore.projects.length === 0}
		<div class="empty-state text-center py-12">
			<svg class="w-24 h-24 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
			</svg>
			<h3 class="text-xl font-medium text-gray-900 mb-2">No projects yet</h3>
			<p class="text-gray-600 mb-6">Get started by creating your first database schema project</p>
			<Button onclick={createNewProject}>
				{#snippet children()}
					Create Your First Project
				{/snippet}
			</Button>
		</div>
	{:else}
		<div class="projects-grid grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			{#each $projectStore.projects as project}
				<div class="project-card bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-md transition-shadow">
					<!-- Project Header -->
					<div class="project-header p-6 border-b border-gray-100">
						<div class="flex items-start justify-between">
							<div class="flex-1 min-w-0">
								<h3 class="text-lg font-medium text-gray-900 truncate">{project.name}</h3>
								<p class="text-sm text-gray-600 mt-1 line-clamp-2">{project.description || 'No description'}</p>
							</div>
							<div class="ml-4 flex-shrink-0">
								<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
									{project.database_type || 'PostgreSQL'}
								</span>
							</div>
						</div>
					</div>

					<!-- Project Stats -->
					<div class="project-stats px-6 py-4 bg-gray-50 border-b border-gray-100">
						<div class="flex items-center justify-between text-sm text-gray-600">
							<div class="flex items-center">
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2v2z" />
								</svg>
								{project.tables?.length || 0} tables
							</div>
							<div class="flex items-center">
								<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								{formatDate(project.updated_at)}
							</div>
						</div>
					</div>

					<!-- Project Actions -->
					<div class="project-actions p-6">
						<div class="flex space-x-3">
							<Button
								variant="outline"
								size="sm"
								onclick={() => openProject(project.id)}
								class="flex-1"
							>
								{#snippet children()}
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
									</svg>
									View
								{/snippet}
							</Button>
							<Button
								size="sm"
								onclick={() => editProject(project.id)}
								class="flex-1"
							>
								{#snippet children()}
									<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
									</svg>
									Edit Schema
								{/snippet}
							</Button>
							<Button
								variant="outline"
								size="sm"
								onclick={() => openDeleteDialog({ id: project.id, name: project.name })}
								class="text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200 hover:border-red-300"
							>
								{#snippet children()}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
									</svg>
								{/snippet}
							</Button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Project Dialog -->
<Dialog bind:open={showCreateDialog} onOpenChange={closeCreateDialog}>
	<div class="space-y-4">
		<div>
			<h2 class="text-lg font-semibold">Create New Project</h2>
			<p class="text-sm text-gray-600">
				Set up a new database schema project
			</p>
		</div>

		<div class="space-y-4">
			<div class="space-y-2">
				<label for="name" class="text-sm font-medium">Project Name</label>
				<Input
					id="name"
					placeholder="My Database Schema"
					bind:value={projectName}
					disabled={isCreating}
					required
				/>
			</div>

			<div class="space-y-2">
				<label for="description" class="text-sm font-medium">Description (optional)</label>
				<Input
					id="description"
					placeholder="Describe your project..."
					bind:value={projectDescription}
					disabled={isCreating}
				/>
			</div>

			<div class="space-y-2">
				<label for="database" class="text-sm font-medium">Database Type</label>
				<Select
					bind:value={databaseType}
					options={databaseOptions}
					disabled={isCreating}
				/>
			</div>
		</div>

		<div class="flex gap-2 pt-4">
			<Button
				variant="outline"
				class="flex-1"
				onclick={closeCreateDialog}
				disabled={isCreating}
			>
				Cancel
			</Button>
			<Button
				class="flex-1"
				onclick={createProject}
				disabled={isCreating || !projectName.trim()}
			>
				{#if isCreating}
					<svg class="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
					Creating...
				{:else}
					Create Project
				{/if}
			</Button>
		</div>
	</div>
</Dialog>

<!-- Delete Project Confirmation Dialog -->
<AlertDialog
	bind:open={showDeleteDialog}
	title="Delete Project"
	description={`Are you sure you want to delete "${projectToDelete?.name}"? This action cannot be undone and all data associated with this project will be permanently deleted.`}
	actionText="Delete Project"
	actionVariant="destructive"
	onAction={confirmDeleteProject}
/>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.project-card:hover {
		transform: translateY(-2px);
	}
</style>