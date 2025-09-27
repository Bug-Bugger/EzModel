<script lang="ts">
	import { useSvelteFlow } from '@xyflow/svelte';
	import { collaborationStore } from '$lib/stores/collaboration';
	import { onMount } from 'svelte';

	// Use SvelteFlow hooks for coordinate conversion - this works inside SvelteFlow context
	const { screenToFlowPosition } = useSvelteFlow();

	let currentCursorPosition: { globalX: number; globalY: number } | null = null;
	let hasNewCursorUpdate = false;
	let cursorUpdateInterval: ReturnType<typeof setInterval> | null = null;
	let stopSendingTimeout: ReturnType<typeof setTimeout> | null = null;

	// Cursor update configuration
	const CURSOR_UPDATE_INTERVAL = 16; // 60fps (16ms)
	const CURSOR_IDLE_DELAY = 200; // Stop sending after 200ms of no movement

	// Handle mouse move for collaboration cursors
	function onMouseMove(event: Event) {
		const mouseEvent = event as MouseEvent;
		// Use SvelteFlow's built-in coordinate conversion via hook
		if (screenToFlowPosition) {
			try {
				// Convert screen coordinates to flow coordinates
				const flowCoords = screenToFlowPosition({
					x: mouseEvent.clientX,
					y: mouseEvent.clientY
				});

				// Validate coordinates are finite numbers
				if (isFinite(flowCoords.x) && isFinite(flowCoords.y)) {
					// Update current position and mark as having new data
					currentCursorPosition = {
						globalX: flowCoords.x,
						globalY: flowCoords.y
					};
					hasNewCursorUpdate = true;

					// Start sending updates if not already started
					startCursorUpdates();

					// Reset the idle timeout
					resetIdleTimeout();
				} else {
					console.warn('Invalid cursor coordinates - not finite:', flowCoords);
				}
			} catch (error) {
				console.warn('Error converting cursor coordinates:', error);
			}
		} else {
			console.warn('SvelteFlow screenToFlowPosition method not available');
		}
	}

	// Start the consistent cursor update interval
	function startCursorUpdates() {
		if (cursorUpdateInterval !== null) return; // Already running

		cursorUpdateInterval = setInterval(() => {
			// Send update if we have new cursor data
			if (hasNewCursorUpdate && currentCursorPosition) {
				collaborationStore.sendCursorPosition(
					currentCursorPosition.globalX,
					currentCursorPosition.globalY
				);
				hasNewCursorUpdate = false; // Mark as sent
			}
		}, CURSOR_UPDATE_INTERVAL);
	}

	// Stop sending cursor updates when idle
	function stopCursorUpdates() {
		if (cursorUpdateInterval !== null) {
			clearInterval(cursorUpdateInterval);
			cursorUpdateInterval = null;
		}
	}

	// Reset the timeout that stops sending when mouse is idle
	function resetIdleTimeout() {
		if (stopSendingTimeout !== null) {
			clearTimeout(stopSendingTimeout);
		}

		stopSendingTimeout = setTimeout(() => {
			stopCursorUpdates();
		}, CURSOR_IDLE_DELAY);
	}

	onMount(() => {
		// Add mouse move listener to the parent canvas container
		const canvasContainer = document.querySelector('.database-canvas');
		if (canvasContainer) {
			canvasContainer.addEventListener('mousemove', onMouseMove);
		} else {
			console.warn('Could not find canvas container for mouse tracking');
		}

		return () => {
			// Clean up cursor update interval
			stopCursorUpdates();

			// Clean up idle timeout
			if (stopSendingTimeout !== null) {
				clearTimeout(stopSendingTimeout);
				stopSendingTimeout = null;
			}

			// Remove event listener
			if (canvasContainer) {
				canvasContainer.removeEventListener('mousemove', onMouseMove);
			}
		};
	});
</script>

<!-- Invisible overlay for mouse tracking inside SvelteFlow -->
<div
	class="absolute inset-0 pointer-events-none"
	style="z-index: 1; background: transparent;"
	role="application"
	aria-label="Mouse tracking overlay for collaboration"
>
	<!-- This div no longer captures events to avoid blocking SvelteFlow interactions -->
</div>