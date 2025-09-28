<script lang="ts">
	import { useSvelteFlow, useNodes, useNodesInitialized } from '@xyflow/svelte';
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	// SvelteFlow hooks - safe to use inside SvelteFlow context
	const { screenToFlowPosition, getViewport } = useSvelteFlow();
	const nodes = useNodes();
	const nodesInitialized = useNodesInitialized();

	// Export coordinate conversion function
	export function convertScreenToFlow(screenX: number, screenY: number) {
		if (screenToFlowPosition) {
			return screenToFlowPosition({ x: screenX, y: screenY });
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
			return getViewport();
		}
		return { x: 0, y: 0, zoom: 1 };
	}

	// Dispatch events when important state changes
	$: if (nodesInitialized) {
		dispatch('nodes-initialized');
	}
</script>

<!-- This component is invisible but provides hook access to parent -->
<div style="display: none;"></div>