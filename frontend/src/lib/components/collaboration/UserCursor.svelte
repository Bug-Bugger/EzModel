<script lang="ts">
	import type { ConnectedUser } from '$lib/stores/collaboration';
	import { useSvelteFlow, type Viewport } from '@xyflow/svelte';
	import { onMount } from 'svelte';

	export let user: ConnectedUser;

	// Use SvelteFlow hooks for coordinate conversion and viewport tracking
	const { flowToScreenPosition, getViewport } = useSvelteFlow();

	let cursorPosition = { x: 0, y: 0 };
	let targetPosition = { x: 0, y: 0 };
	let smoothingInterval: ReturnType<typeof setInterval> | null = null;
	let currentViewport: Viewport | null = null;
	let containerRect: DOMRect | null = null;

	// Smoothing configuration
	const SMOOTHING_FACTOR = 0.3; // How much to move towards target each frame (increased for responsiveness)
	const SMOOTHING_FPS = 60; // Target smoothing frame rate
	const SMOOTHING_INTERVAL = 1000 / SMOOTHING_FPS; // ~16ms

	// Update target position when user cursor data changes
	$: if (user.cursor) {
		updateTargetPosition();
	}

	// Track viewport changes and update cursor position accordingly
	$: if (user.cursor && getViewport) {
		// Get current viewport to make this reactive to viewport changes
		const newViewport = getViewport();
		// Only update if viewport actually changed and is valid (avoid infinite loops)
		if (
			newViewport &&
			newViewport.zoom !== undefined &&
			JSON.stringify(newViewport) !== JSON.stringify(currentViewport)
		) {
			currentViewport = newViewport;
			updateTargetPosition();
		}
	}

	function updateContainerRect() {
		// Get the SvelteFlow container's position relative to the viewport
		const flowContainer = document.querySelector('.svelte-flow');
		if (flowContainer) {
			containerRect = flowContainer.getBoundingClientRect();
		}
	}

	function updateTargetPosition() {
		if (!user.cursor) return;

		if (!flowToScreenPosition) {
			console.warn('SvelteFlow flowToScreenPosition not available');
			return;
		}

		// Update container rect to account for layout changes
		updateContainerRect();

		try {
			// Convert flow coordinates (received from other user) to screen coordinates
			const screenCoords = flowToScreenPosition({
				x: user.cursor.x,
				y: user.cursor.y
			});

			// Validate that the converted coordinates are valid
			if (!isFinite(screenCoords.x) || !isFinite(screenCoords.y)) {
				console.warn('Invalid screen coordinates after conversion:', screenCoords);
				return;
			}

			// Adjust for container offset - flowToScreenPosition returns viewport-relative coords
			// but we need container-relative coords since cursor uses absolute positioning
			let adjustedX = screenCoords.x;
			let adjustedY = screenCoords.y;

			if (containerRect) {
				adjustedX = screenCoords.x - containerRect.left;
				adjustedY = screenCoords.y - containerRect.top;
			}

			targetPosition = {
				x: adjustedX,
				y: adjustedY
			};

			// Start smoothing if not already running
			startSmoothing();
		} catch (error) {
			console.warn('Error converting cursor coordinates:', error);
		}
	}

	function startSmoothing() {
		if (smoothingInterval !== null) return; // Already running

		smoothingInterval = setInterval(() => {
			// Calculate distance to target
			const dx = targetPosition.x - cursorPosition.x;
			const dy = targetPosition.y - cursorPosition.y;
			const distance = Math.sqrt(dx * dx + dy * dy);

			// If we're close enough, snap to target and stop smoothing
			if (distance < 0.5) {
				cursorPosition = { ...targetPosition };
				stopSmoothing();
				return;
			}

			// Smooth movement towards target using lerp
			cursorPosition = {
				x: cursorPosition.x + dx * SMOOTHING_FACTOR,
				y: cursorPosition.y + dy * SMOOTHING_FACTOR
			};
		}, SMOOTHING_INTERVAL);
	}

	function stopSmoothing() {
		if (smoothingInterval !== null) {
			clearInterval(smoothingInterval);
			smoothingInterval = null;
		}
	}

	onMount(() => {
		// Initialize viewport tracking
		if (getViewport) {
			const initialViewport = getViewport();
			if (initialViewport && initialViewport.zoom !== undefined) {
				currentViewport = initialViewport;
			}
		}

		// Initialize container rect
		updateContainerRect();
		updateTargetPosition();

		// Update container rect on window resize or layout changes
		function handleResize() {
			updateContainerRect();
			updateTargetPosition();
		}

		window.addEventListener('resize', handleResize);

		return () => {
			// Clean up smoothing interval on unmount
			stopSmoothing();
			window.removeEventListener('resize', handleResize);
		};
	});

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
			'#6366f1' // indigo
		];

		// Simple hash function to get consistent color
		let hash = 0;
		for (let i = 0; i < userId.length; i++) {
			hash = ((hash << 5) - hash + userId.charCodeAt(i)) & 0xffffffff;
		}
		return colors[Math.abs(hash) % colors.length];
	}

	$: userColor = getUserColor(user.id);
	$: isActive = user.cursor && Date.now() - user.cursor.timestamp < 5000; // Show cursor for 5 seconds after last movement
</script>

{#if isActive && user.cursor}
	<div
		class="user-cursor absolute pointer-events-none z-50"
		style="left: {cursorPosition.x}px; top: {cursorPosition.y}px; color: {userColor}; transform: translateZ(0);"
	>
		<!-- Cursor Arrow -->
		<svg
			class="cursor-arrow w-6 h-6 transform -translate-x-1 -translate-y-1"
			fill="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				d="M12 2L2 7L3 8L12 5L21 8L22 7L12 2ZM12 5L3 8V18C3 19.1 3.9 20 5 20H19C20.1 20 21 19.1 21 18V8L12 5Z"
			/>
		</svg>

		<!-- User Name Label -->
		<div
			class="user-label ml-6 -mt-1 px-2 py-1 rounded text-xs font-medium text-white shadow-lg max-w-24 truncate"
			style="background-color: {userColor}"
		>
			{user.username || 'Unknown User'}
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
