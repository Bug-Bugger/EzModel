<script lang="ts">
	import { useSvelteFlow, useNodesInitialized } from '@xyflow/svelte';

	// Removed deprecated createEventDispatcher - not needed for current functionality

	// SvelteFlow hooks - safe to use inside SvelteFlow context
	const { screenToFlowPosition, getViewport } = useSvelteFlow();
	const nodesInitialized = useNodesInitialized();

	// Export coordinate conversion function
	export function convertScreenToFlow(screenX: number, screenY: number) {
		if (screenToFlowPosition) {
			try {
				const result = screenToFlowPosition({ x: screenX, y: screenY });
				// Ensure the result is valid before returning
				if (result && isFinite(result.x) && isFinite(result.y)) {
					return result;
				}
			} catch (error) {
				console.warn('Error in screenToFlowPosition:', error);
			}
		}
		return { x: screenX, y: screenY };
	}

	// Export nodes initialization status
	export function getNodesInitialized() {
		return nodesInitialized;
	}

	// Export current viewport
	export function getCurrentViewport() {
		if (getViewport) {
			const viewport = getViewport();
			// Return the viewport only if it's valid, otherwise return default
			if (viewport && viewport.zoom !== undefined) {
				return viewport;
			}
		}
		return { x: 0, y: 0, zoom: 1 };
	}

	// Removed dispatch functionality - not currently needed
</script>

<!-- This component is invisible but provides hook access to parent -->
<div style="display: none;"></div>
