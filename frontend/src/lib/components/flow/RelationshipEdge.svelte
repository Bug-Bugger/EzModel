<script lang="ts">
	import { BaseEdge, getSmoothStepPath, type EdgeProps } from '@xyflow/svelte';
	import type { RelationshipEdge } from '$lib/stores/flow';

	let {
		sourceX,
		sourceY,
		targetX,
		targetY,
		sourcePosition,
		targetPosition,
		data,
		markerEnd,
		...restProps
	}: EdgeProps = $props();

	// Cast data to our relationship edge data type
	const relationshipData = data as RelationshipEdge['data'] | undefined;
	const relationshipType = relationshipData?.relation_type || 'one_to_many';

	// Generate path using getSmoothStepPath
	const pathResult = $derived(
		getSmoothStepPath({
			sourceX,
			sourceY,
			sourcePosition,
			targetX,
			targetY,
			targetPosition,
			borderRadius: 10
		})
	);

	const path = $derived(pathResult[0]); // getSmoothStepPath returns array, first element is the path

	// Label position and text
	const labelX = $derived((sourceX + targetX) / 2);
	const labelY = $derived((sourceY + targetY) / 2);

	// Get relationship type symbols - handle both hyphenated and underscore formats
	function getRelationshipSymbol(type: string) {
		switch (type) {
			case 'one-to-one':
			case 'one_to_one':
				return '1:1';
			case 'one-to-many':
			case 'one_to_many':
				return '1:N';
			case 'many-to-many':
			case 'many_to_many':
				return 'N:M';
			default:
				return '1:N';
		}
	}

	function getRelationshipColor(type: string) {
		switch (type) {
			case 'one-to-one':
			case 'one_to_one':
				return '#10b981'; // green
			case 'one-to-many':
			case 'one_to_many':
				return '#3b82f6'; // blue
			case 'many-to-many':
			case 'many_to_many':
				return '#f59e0b'; // yellow
			default:
				return '#64748b'; // gray
		}
	}
</script>

<!-- Base edge path -->
<BaseEdge {path} style={`stroke: ${getRelationshipColor(relationshipType)}; stroke-width: 2px;`} />

<!-- Custom relationship type label -->
<foreignObject x={labelX - 20} y={labelY - 10} width="40" height="20" class="edge-label">
	<div
		class="relationship-label"
		style={`
			background: white;
			border: 1px solid #d1d5db;
			border-radius: 4px;
			padding: 2px 6px;
			font-size: 11px;
			font-weight: 500;
			color: ${getRelationshipColor(relationshipType)};
			box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
			text-align: center;
			white-space: nowrap;
		`}
	>
		{getRelationshipSymbol(relationshipType)}
	</div>
</foreignObject>
