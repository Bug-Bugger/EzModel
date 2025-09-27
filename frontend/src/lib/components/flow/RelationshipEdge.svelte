<script lang="ts">
	import { getSmoothStepPath } from '@xyflow/svelte';
	import type { RelationshipEdge } from '$lib/stores/flow';

	export let id: string;
	export let sourceX: number;
	export let sourceY: number;
	export let targetX: number;
	export let targetY: number;
	export let sourcePosition: any;
	export let targetPosition: any;
	export let data: RelationshipEdge['data'];
	export let selected: boolean = false;

	$: pathResult = getSmoothStepPath({
		sourceX,
		sourceY,
		sourcePosition,
		targetX,
		targetY,
		targetPosition,
		borderRadius: 10
	});

	$: path = pathResult[0]; // getSmoothStepPath returns array, first element is the path
	$: labelX = (sourceX + targetX) / 2;
	$: labelY = (sourceY + targetY) / 2;

	// Get relationship type symbols
	function getRelationshipSymbol(type: string) {
		switch (type) {
			case 'one-to-one': return '1:1';
			case 'one-to-many': return '1:N';
			case 'many-to-many': return 'N:M';
			default: return '1:N';
		}
	}

	function getRelationshipColor(type: string) {
		switch (type) {
			case 'one-to-one': return '#10b981'; // green
			case 'one-to-many': return '#3b82f6'; // blue
			case 'many-to-many': return '#f59e0b'; // yellow
			default: return '#64748b'; // gray
		}
	}
</script>

<!-- Edge Path -->
<path
	{id}
	d={path}
	stroke={selected ? '#3b82f6' : getRelationshipColor(data.type)}
	stroke-width={selected ? 3 : 2}
	fill="none"
	class="relationship-edge"
	marker-end="url(#arrowhead-{data.type})"
/>

<!-- Edge Label -->
<div
	class="relationship-label absolute pointer-events-none"
	style="transform: translate(-50%, -50%) translate({labelX}px, {labelY}px)"
>
	<div
		class="label-content bg-white border border-gray-300 rounded px-2 py-1 text-xs font-medium shadow-sm"
		style="color: {getRelationshipColor(data.type)}"
	>
		{getRelationshipSymbol(data.type)}
	</div>
</div>

<!-- Arrow Markers -->
<defs>
	<marker
		id="arrowhead-{data.type}"
		markerWidth="10"
		markerHeight="7"
		refX="9"
		refY="3.5"
		orient="auto"
	>
		<polygon
			points="0 0, 10 3.5, 0 7"
			fill={selected ? '#3b82f6' : getRelationshipColor(data.type)}
		/>
	</marker>
</defs>

<style>
	.relationship-edge {
		cursor: pointer;
		transition: stroke-width 0.2s, stroke 0.2s;
	}

	.relationship-edge:hover {
		stroke-width: 3;
	}

	.relationship-label {
		font-family: 'Inter', sans-serif;
	}

	.label-content {
		min-width: 24px;
		text-align: center;
	}
</style>