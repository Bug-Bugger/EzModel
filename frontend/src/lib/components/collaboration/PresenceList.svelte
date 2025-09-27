<script lang="ts">
	import { collaborationStore } from '$lib/stores/collaboration';

	// Generate consistent avatar colors
	function getUserColor(userId: string): string {
		const colors = [
			'bg-blue-500', 'bg-green-500', 'bg-yellow-500', 'bg-red-500', 'bg-purple-500',
			'bg-cyan-500', 'bg-orange-500', 'bg-lime-500', 'bg-pink-500', 'bg-indigo-500'
		];

		let hash = 0;
		for (let i = 0; i < userId.length; i++) {
			hash = ((hash << 5) - hash + userId.charCodeAt(i)) & 0xffffffff;
		}
		return colors[Math.abs(hash) % colors.length];
	}

	// Get user initials for avatar
	function getUserInitials(name: string): string {
		return name
			.split(' ')
			.map(word => word.charAt(0))
			.join('')
			.slice(0, 2)
			.toUpperCase();
	}

	// Check if user is currently active (moved cursor recently)
	function isUserActive(user: any): boolean {
		if (!user.cursor) return false;
		return (Date.now() - user.cursor.timestamp) < 10000; // Active within last 10 seconds
	}
</script>

<div class="presence-list flex items-center space-x-2">
	{#if $collaborationStore.isConnected && $collaborationStore.connectedUsers.length > 0}
		<!-- Connected Users Avatars -->
		<div class="user-avatars flex -space-x-2">
			{#each $collaborationStore.connectedUsers as user}
				<div
					class="user-avatar relative"
					title={user.name}
				>
					<!-- Avatar -->
					{#if user.avatar}
						<img
							src={user.avatar}
							alt={user.name}
							class="w-8 h-8 rounded-full border-2 border-white shadow-sm"
						/>
					{:else}
						<div
							class="w-8 h-8 rounded-full border-2 border-white shadow-sm flex items-center justify-center text-white text-xs font-medium {getUserColor(user.id)}"
						>
							{getUserInitials(user.name)}
						</div>
					{/if}

					<!-- Activity Indicator -->
					{#if isUserActive(user)}
						<div class="activity-indicator absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-400 border-2 border-white rounded-full"></div>
					{/if}
				</div>
			{/each}
		</div>

		<!-- More Users Indicator -->
		{#if $collaborationStore.connectedUsers.length > 5}
			<div class="more-users text-xs text-gray-500 ml-2">
				+{$collaborationStore.connectedUsers.length - 5} more
			</div>
		{/if}

		<!-- Users Dropdown (when hovering) -->
		<div class="users-dropdown relative group">
			<button
				class="users-toggle p-1 rounded hover:bg-gray-100 transition-colors"
				aria-label="Show connected users"
			>
				<svg class="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>
			</button>

			<!-- Dropdown Content -->
			<div class="dropdown-content absolute right-0 top-full mt-2 bg-white border border-gray-200 rounded-lg shadow-lg p-2 min-w-48 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50">
				<div class="text-xs font-medium text-gray-900 mb-2 px-2">Online Users</div>

				{#each $collaborationStore.connectedUsers as user}
					<div class="user-item flex items-center space-x-2 px-2 py-1 rounded hover:bg-gray-50">
						<!-- Avatar -->
						{#if user.avatar}
							<img
								src={user.avatar}
								alt={user.name}
								class="w-6 h-6 rounded-full"
							/>
						{:else}
							<div
								class="w-6 h-6 rounded-full flex items-center justify-center text-white text-xs font-medium {getUserColor(user.id)}"
							>
								{getUserInitials(user.name)}
							</div>
						{/if}

						<!-- User Info -->
						<div class="flex-1 min-w-0">
							<div class="text-sm font-medium text-gray-900 truncate">{user.name}</div>
							<div class="text-xs text-gray-500 truncate">{user.email}</div>
						</div>

						<!-- Activity Status -->
						<div class="flex items-center">
							{#if isUserActive(user)}
								<div class="w-2 h-2 bg-green-400 rounded-full" title="Active"></div>
							{:else}
								<div class="w-2 h-2 bg-gray-300 rounded-full" title="Idle"></div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.user-avatar {
		position: relative;
		transition: transform 0.2s;
	}

	.user-avatar:hover {
		transform: translateY(-2px);
		z-index: 10;
	}

	.activity-indicator {
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.users-dropdown .dropdown-content {
		transform-origin: top right;
	}

	.presence-list {
		font-family: 'Inter', sans-serif;
	}
</style>