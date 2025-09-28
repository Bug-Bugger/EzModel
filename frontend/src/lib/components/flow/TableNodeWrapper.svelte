<script lang="ts">
	import TableNode from './TableNode.svelte';
	import { flowStore } from '$lib/stores/flow';
	import { designerStore } from '$lib/stores/designer';
	import type { TableNode as TableNodeType } from '$lib/stores/flow';

	export let data: TableNodeType['data'];
	export let selected: boolean = false;

	// Handle the addField event from TableNode
	function handleAddField(event: any) {
		const { tableId, tableName } = event.detail;

		// Find the table node
		const tableNode = { id: tableId, data, type: 'table', position: data.position } as TableNodeType;

		// Select the table and open property panel
		flowStore.selectNode(tableNode);
		designerStore.openPropertyPanel('table', tableNode);
	}
</script>

<TableNode {data} {selected} on:addField={handleAddField} />