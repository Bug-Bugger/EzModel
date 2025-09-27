<script lang="ts">
	import { onMount } from 'svelte';
	import Dialog from '$lib/components/ui/dialog.svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import { userService } from '$lib/services/user';
	import { projectStore } from '$lib/stores/project';
	import type { User } from '$lib/types/models';

	type Props = {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
	};

	let { open = $bindable(false), onOpenChange }: Props = $props();

	let searchQuery = $state('');
	let users = $state<User[]>([]);
	let isLoading = $state(false);
	let isAddingCollaborator = $state(false);
	let selectedUser = $state<User | null>(null);

	const currentProject = $derived($projectStore.currentProject);
	const existingCollaboratorIds = $derived(currentProject?.collaborators?.map(c => c.id) || []);

	// Filter users based on search query and exclude existing collaborators and owner
	const filteredUsers = $derived(users
		.filter(user => {
			const matchesSearch = searchQuery === '' ||
				(user.email || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
				(user.username || '').toLowerCase().includes(searchQuery.toLowerCase());

			const isNotOwner = user.id !== currentProject?.owner_id;
			const isNotCollaborator = !existingCollaboratorIds.includes(user.id);

			return matchesSearch && isNotOwner && isNotCollaborator;
		})
		.slice(0, 10) // Limit to 10 results
	);

	onMount(async () => {
		await loadUsers();
	});

	async function loadUsers() {
		isLoading = true;
		try {
			users = await userService.getAllUsers();
		} catch (error) {
			console.error('Failed to load users:', error);
		} finally {
			isLoading = false;
		}
	}

	async function addCollaborator(user: User) {
		if (!user || isAddingCollaborator) return;

		isAddingCollaborator = true;
		try {
			await projectStore.addCollaborator(user.id);
			close();
			// Reset form
			searchQuery = '';
			selectedUser = null;
		} catch (error) {
			console.error('Failed to add collaborator:', error);
			alert('Failed to add collaborator. Please try again.');
		} finally {
			isAddingCollaborator = false;
		}
	}

	function close() {
		open = false;
		onOpenChange?.(false);
		searchQuery = '';
		selectedUser = null;
	}

	function selectUser(user: User) {
		selectedUser = user;
		searchQuery = user.username || '';
	}
</script>

<Dialog bind:open {onOpenChange}>
	{#snippet children()}
		<div class="space-y-6">
			<div>
				<h2 class="text-lg font-semibold text-gray-900">Add Collaborator</h2>
				<p class="text-sm text-gray-600 mt-1">
					Search for users to add as collaborators to this project
				</p>
			</div>

			<div class="space-y-4">
				<!-- Search Input -->
				<div>
					<label for="user-search" class="block text-sm font-medium text-gray-700 mb-2">
						Search Users
					</label>
					<Input
						id="user-search"
						bind:value={searchQuery}
						placeholder="Search by name, email, or username..."
						class="w-full"
						disabled={isLoading}
					/>
				</div>

				<!-- Loading State -->
				{#if isLoading}
					<div class="flex items-center justify-center py-8">
						<svg class="animate-spin h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						<span class="ml-2 text-gray-600">Loading users...</span>
					</div>
				{:else}
					<!-- User Results -->
					<div class="space-y-2 max-h-64 overflow-y-auto">
						{#if searchQuery && filteredUsers.length === 0}
							<div class="text-center py-6 text-gray-500">
								<svg class="w-12 h-12 mx-auto mb-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
								</svg>
								<p class="text-sm">No users found</p>
							</div>
						{:else if searchQuery}
							{#each filteredUsers as user (user.id)}
								<button
									class="w-full p-3 text-left rounded-lg border hover:bg-gray-50 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500"
									onclick={() => selectUser(user)}
								>
									<div class="flex items-center space-x-3">
										<div class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
											{(user.username || 'U').charAt(0).toUpperCase()}
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium text-gray-900 truncate">{user.username || 'Unknown User'}</p>
											<p class="text-xs text-gray-500 truncate">{user.email || 'No email'}</p>
										</div>
									</div>
								</button>
							{/each}
						{:else}
							<div class="text-center py-6 text-gray-500">
								<svg class="w-12 h-12 mx-auto mb-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
								</svg>
								<p class="text-sm">Start typing to search for users</p>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Actions -->
			<div class="flex justify-end space-x-3 pt-4 border-t">
				<Button variant="outline" onclick={close} disabled={isAddingCollaborator}>
					Cancel
				</Button>
				<Button
					onclick={() => selectedUser && addCollaborator(selectedUser)}
					disabled={!selectedUser || isAddingCollaborator}
					class="min-w-24"
				>
					{#if isAddingCollaborator}
						<svg class="animate-spin h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Adding...
					{:else}
						Add Collaborator
					{/if}
				</Button>
			</div>
		</div>
	{/snippet}
</Dialog>