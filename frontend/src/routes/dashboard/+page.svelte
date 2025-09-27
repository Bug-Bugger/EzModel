<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { projectStore } from '$lib/stores/project';
	import { uiStore } from '$lib/stores/ui';
	import { projectService } from '$lib/services/project';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import Select from '$lib/components/ui/select.svelte';
	import Dialog from '$lib/components/ui/dialog.svelte';
	import AlertDialog from '$lib/components/ui/alert-dialog.svelte';
	import { Plus, Database, Calendar, MoreHorizontal, Trash2, Edit } from 'lucide-svelte';
	import type { CreateProjectRequest } from '$lib/types/models';

	// Remove props, use stores directly

	// Project creation form
	let showCreateDialog = false;
	let projectName = '';
	let projectDescription = '';
	let databaseType = 'postgresql';
	let isCreating = false;

	// Delete confirmation dialog
	let showDeleteDialog = false;
	let projectToDelete: { id: string; name: string } | null = null;

	// Database type options
	const databaseOptions = [
		{ value: 'postgresql', label: 'PostgreSQL' },
		{ value: 'mysql', label: 'MySQL' },
		{ value: 'sqlite', label: 'SQLite' },
		{ value: 'sqlserver', label: 'SQL Server' }
	];

	// Redirect if not authenticated
	onMount(async () => {
		if (!$authStore.isAuthenticated) {
			goto('/login');
			return;
		}

		// Load user's projects
		try {
			await projectStore.loadProjects();
		} catch (error) {
			uiStore.error('Failed to load projects', 'Please try refreshing the page');
		}
	});

	function openCreateDialog() {
		showCreateDialog = true;
		projectName = '';
		projectDescription = '';
		databaseType = 'postgresql';
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
		} catch (error: any) {
			uiStore.error('Failed to delete project', error.message);
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function getDatabaseIcon(dbType: string) {
		return Database; // Using single icon for simplicity
	}
</script>

<svelte:head>
	<title>Dashboard - EzModel</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<!-- Header -->
	<div class="flex items-center justify-between mb-8">
		<div>
			<h1 class="text-3xl font-bold">Projects</h1>
			<p class="text-muted-foreground mt-2">
				Manage your database schema projects
			</p>
		</div>
		<Button onclick={openCreateDialog}>
			<Plus class="mr-2 h-4 w-4" />
			New Project
		</Button>
	</div>

	<!-- Projects Grid -->
	{#if $projectStore.isLoading}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each Array(6) as _}
				<Card class="p-6">
					<div class="animate-pulse">
						<div class="h-4 bg-muted rounded w-3/4 mb-2"></div>
						<div class="h-3 bg-muted rounded w-1/2 mb-4"></div>
						<div class="flex items-center justify-between">
							<div class="h-3 bg-muted rounded w-1/3"></div>
							<div class="h-6 w-6 bg-muted rounded"></div>
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{:else if $projectStore.projects.length === 0}
		<div class="text-center py-16">
			<Database class="mx-auto h-16 w-16 text-muted-foreground mb-4" />
			<h2 class="text-2xl font-semibold mb-2">No projects yet</h2>
			<p class="text-muted-foreground mb-6">
				Create your first database schema project to get started
			</p>
			<Button onclick={openCreateDialog}>
				<Plus class="mr-2 h-4 w-4" />
				Create Your First Project
			</Button>
		</div>
	{:else}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each $projectStore.projects as project (project.id)}
				<Card class="p-6 hover:shadow-lg transition-shadow group">
					<div class="flex items-start justify-between mb-4">
						<div class="flex items-center gap-3">
							<div class="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
								<svelte:component this={getDatabaseIcon(project.database_type)} class="h-5 w-5 text-primary" />
							</div>
							<div>
								<h3 class="font-semibold text-lg">{project.name}</h3>
								<p class="text-sm text-muted-foreground capitalize">
									{project.database_type}
								</p>
							</div>
						</div>
						<div class="opacity-0 group-hover:opacity-100 transition-opacity">
							<Button variant="ghost" size="icon" onclick={() => openDeleteDialog({ id: project.id, name: project.name })}>
								<Trash2 class="h-4 w-4" />
							</Button>
						</div>
					</div>

					{#if project.description}
						<p class="text-sm text-muted-foreground mb-4 line-clamp-2">
							{project.description}
						</p>
					{/if}

					<div class="flex items-center justify-between pt-4 border-t">
						<div class="flex items-center gap-2 text-xs text-muted-foreground">
							<Calendar class="h-3 w-3" />
							{formatDate(project.updated_at)}
						</div>
						<Button variant="outline" size="sm" onclick={() => uiStore.info('Editor coming soon!')}>
							<Edit class="mr-2 h-3 w-3" />
							Edit
						</Button>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Project Dialog -->
<Dialog bind:open={showCreateDialog} onOpenChange={closeCreateDialog}>
	<div class="space-y-4">
		<div>
			<h2 class="text-lg font-semibold">Create New Project</h2>
			<p class="text-sm text-muted-foreground">
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
					<div class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent mr-2"></div>
					Creating...
				{:else}
					Create Project
				{/if}
			</Button>
		</div>
	</div>
</Dialog>

<!-- Delete Project Alert Dialog -->
<AlertDialog
	bind:open={showDeleteDialog}
	title="Delete Project"
	description={`Are you sure you want to delete "${projectToDelete?.name}"? This action cannot be undone and all data associated with this project will be permanently deleted.`}
	actionText="Delete Project"
	actionVariant="destructive"
	onAction={confirmDeleteProject}
/>