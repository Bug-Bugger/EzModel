<script lang="ts">
	import type { ConnectedUser } from '$lib/stores/collaboration';

	export let user: ConnectedUser;

	// Generate a consistent color for each user based on their ID
	function getUserColor(userId: string): string {
		const colors = [
			'#3b82f6', // blue
			'#10b981', // green
			'#f59e0b', // yellow
			'#ef4444', // red
			'#8b5cf6', // purple
			'#06b6d4', // cyan
			'#f97316', // orange
			'#84cc16', // lime
			'#ec4899', // pink
			'#6366f1'  // indigo
		];

		// Simple hash function to get consistent color
		let hash = 0;
		for (let i = 0; i < userId.length; i++) {
			hash = ((hash << 5) - hash + userId.charCodeAt(i)) & 0xffffffff;
		}
		return colors[Math.abs(hash) % colors.length];
	}

	$: userColor = getUserColor(user.id);
	$: isActive = user.cursor && (Date.now() - user.cursor.timestamp) < 5000; // Show cursor for 5 seconds after last movement
</script>

{#if isActive && user.cursor}
	<div
		class="user-cursor absolute pointer-events-none z-50 transition-all duration-100 ease-out"
		style="left: {user.cursor.x}px; top: {user.cursor.y}px; color: {userColor}"
	>
		<!-- Cursor Arrow -->
		<svg
			class="cursor-arrow w-6 h-6 transform -translate-x-1 -translate-y-1"
			fill="currentColor"
			viewBox="0 0 24 24"
		>
			<path d="M12 2L2 7L3 8L12 5L21 8L22 7L12 2ZM12 5L3 8V18C3 19.1 3.9 20 5 20H19C20.1 20 21 19.1 21 18V8L12 5Z"/>
		</svg>

		<!-- User Name Label -->
		<div
			class="user-label ml-6 -mt-1 px-2 py-1 rounded text-xs font-medium text-white shadow-lg max-w-24 truncate"
			style="background-color: {userColor}"
		>
			{user.name}
		</div>
	</div>
{/if}

<style>
	.user-cursor {
		transform-origin: 0 0;
		filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
	}

	.cursor-arrow {
		transform: rotate(-45deg);
	}

	.user-label {
		font-family: 'Inter', sans-serif;
		white-space: nowrap;
	}
</style>