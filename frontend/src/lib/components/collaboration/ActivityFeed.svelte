<script lang="ts">
	import { collaborationStore } from '$lib/stores/collaboration';

	// Format timestamp to relative time
	function formatTime(timestamp: number): string {
		const now = Date.now();
		const diff = now - timestamp;

		if (diff < 60000) {
			return 'just now';
		} else if (diff < 3600000) {
			const minutes = Math.floor(diff / 60000);
			return `${minutes}m ago`;
		} else if (diff < 86400000) {
			const hours = Math.floor(diff / 3600000);
			return `${hours}h ago`;
		} else {
			const days = Math.floor(diff / 86400000);
			return `${days}d ago`;
		}
	}

	// Get icon for activity type
	function getActivityIcon(type: string): string {
		switch (type) {
			case 'table_create': return '🗂️';
			case 'table_update': return '✏️';
			case 'table_delete': return '🗑️';
			case 'field_create': return '📝';
			case 'field_update': return '🔧';
			case 'field_delete': return '❌';
			case 'relationship_create': return '🔗';
			case 'relationship_delete': return '💔';
			case 'user_joined': return '👋';
			default: return '📄';
		}
	}

	// Get color class for activity type
	function getActivityColor(type: string): string {
		switch (type) {
			case 'table_create':
			case 'field_create':
			case 'relationship_create':
			case 'user_joined':
				return 'text-green-600 bg-green-50 border-green-200';
			case 'table_update':
			case 'field_update':
				return 'text-blue-600 bg-blue-50 border-blue-200';
			case 'table_delete':
			case 'field_delete':
			case 'relationship_delete':
				return 'text-red-600 bg-red-50 border-red-200';
			default:
				return 'text-gray-600 bg-gray-50 border-gray-200';
		}
	}

	// Generate consistent user color
	function getUserColor(userId: string): string {
		const colors = [
			'text-blue-600', 'text-green-600', 'text-yellow-600', 'text-red-600', 'text-purple-600',
			'text-cyan-600', 'text-orange-600', 'text-lime-600', 'text-pink-600', 'text-indigo-600'
		];

		let hash = 0;
		for (let i = 0; i < userId.length; i++) {
			hash = ((hash << 5) - hash + userId.charCodeAt(i)) & 0xffffffff;
		}
		return colors[Math.abs(hash) % colors.length];
	}

	// Clear all activity
	function clearActivity() {
		collaborationStore.clearActivity();
	}
</script>

<div class="activity-feed h-full flex flex-col">
	<!-- Header -->
	<div class="activity-header flex items-center justify-between p-4 border-b border-gray-200">
		<h3 class="text-sm font-medium text-gray-900">Recent Activity</h3>
		{#if $collaborationStore.activityEvents.length > 0}
			<button
				on:click={clearActivity}
				class="text-xs text-gray-500 hover:text-gray-700 hover:underline"
			>
				Clear
			</button>
		{/if}
	</div>

	<!-- Activity List -->
	<div class="activity-list flex-1 overflow-y-auto p-4">
		{#if $collaborationStore.activityEvents.length === 0}
			<div class="empty-state text-center py-8">
				<svg class="w-12 h-12 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<p class="text-sm text-gray-500">No recent activity</p>
				<p class="text-xs text-gray-400 mt-1">Collaboration events will appear here</p>
			</div>
		{:else}
			<div class="space-y-3">
				{#each $collaborationStore.activityEvents as event}
					<div class="activity-item border rounded-lg p-3 {getActivityColor(event.type)}">
						<div class="flex items-start space-x-3">
							<!-- Activity Icon -->
							<div class="activity-icon flex-shrink-0 w-6 h-6 flex items-center justify-center text-sm">
								{getActivityIcon(event.type)}
							</div>

							<!-- Activity Content -->
							<div class="activity-content flex-1 min-w-0">
								<div class="activity-message text-sm">
									<span class="user-name font-medium {getUserColor(event.userId)}">
										{event.userName}
									</span>
									<span class="activity-text text-gray-700 ml-1">
										{event.message}
									</span>
								</div>

								<!-- Timestamp -->
								<div class="activity-time text-xs text-gray-500 mt-1">
									{formatTime(event.timestamp)}
								</div>

								<!-- Additional Data (if available) -->
								{#if event.data && event.type.includes('table')}
									<div class="activity-data text-xs bg-white bg-opacity-50 rounded px-2 py-1 mt-2">
										<span class="font-mono text-gray-600">
											{event.data.name || 'Unknown'}
										</span>
									</div>
								{/if}
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Live Indicator -->
	{#if $collaborationStore.isConnected}
		<div class="live-indicator flex items-center justify-center p-2 border-t border-gray-200 bg-green-50">
			<div class="flex items-center space-x-2 text-xs text-green-600">
				<div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
				<span>Live updates active</span>
			</div>
		</div>
	{/if}
</div>

<style>
	.activity-feed {
		font-family: 'Inter', sans-serif;
	}

	.activity-item {
		transition: all 0.2s;
		animation: slideIn 0.3s ease-out;
	}

	.activity-item:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 8px -2px rgba(0, 0, 0, 0.1);
	}

	@keyframes slideIn {
		from {
			opacity: 0;
			transform: translateX(10px);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	.activity-list {
		scrollbar-width: thin;
		scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
	}

	.activity-list::-webkit-scrollbar {
		width: 4px;
	}

	.activity-list::-webkit-scrollbar-track {
		background: transparent;
	}

	.activity-list::-webkit-scrollbar-thumb {
		background-color: rgba(156, 163, 175, 0.5);
		border-radius: 2px;
	}
</style>