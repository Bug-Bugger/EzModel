<script lang="ts">
	import { collaborationStore } from '$lib/stores/collaboration';

	// Connection status indicators
	function getStatusColor(status: string): string {
		switch (status) {
			case 'connected':
				return 'text-green-600 bg-green-100';
			case 'connecting':
				return 'text-yellow-600 bg-yellow-100';
			case 'disconnected':
				return 'text-gray-600 bg-gray-100';
			case 'error':
				return 'text-red-600 bg-red-100';
			default:
				return 'text-gray-600 bg-gray-100';
		}
	}

	function getStatusText(status: string): string {
		switch (status) {
			case 'connected':
				return 'Connected';
			case 'connecting':
				return 'Connecting...';
			case 'disconnected':
				return 'Disconnected';
			case 'error':
				return 'Connection Error';
			default:
				return 'Unknown';
		}
	}

	function getStatusIcon(status: string): string {
		switch (status) {
			case 'connected':
				return '●';
			case 'connecting':
				return '○';
			case 'disconnected':
				return '●';
			case 'error':
				return '⚠';
			default:
				return '○';
		}
	}
</script>

<div class="collaboration-status flex items-center space-x-3">
	<!-- Connection Status -->
	<div class="flex items-center space-x-2">
		<div
			class="status-indicator px-2 py-1 rounded-full text-xs font-medium {getStatusColor(
				$collaborationStore.connectionStatus
			)}"
		>
			<span class="status-icon mr-1">{getStatusIcon($collaborationStore.connectionStatus)}</span>
			{getStatusText($collaborationStore.connectionStatus)}
		</div>

		{#if $collaborationStore.lastError}
			<div
				class="error-message text-xs text-red-600 max-w-48 truncate"
				title={$collaborationStore.lastError}
			>
				{$collaborationStore.lastError}
			</div>
		{/if}
	</div>

	<!-- User Count -->
	{#if $collaborationStore.isConnected}
		<div class="user-count flex items-center space-x-1 text-sm text-gray-600">
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
				/>
			</svg>
			<span>
				{$collaborationStore.connectedUsers.length}
				{$collaborationStore.connectedUsers.length === 1 ? 'user' : 'users'} online
			</span>
		</div>
	{/if}
</div>

<style>
	.status-indicator {
		min-width: fit-content;
		display: flex;
		align-items: center;
	}

	.status-icon {
		line-height: 1;
	}

	.collaboration-status {
		font-family: 'Inter', sans-serif;
	}
</style>
