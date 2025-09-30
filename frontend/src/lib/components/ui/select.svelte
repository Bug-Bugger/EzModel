<script lang="ts">
	import { cn } from '$lib/utils/cn';

	type Props = {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		class?: string;
		options: { value: string; label: string }[];
		onchange?: (value: string) => void;
	};

	let {
		value = $bindable(''),
		placeholder = 'Select an option...',
		disabled = false,
		required = false,
		class: className,
		options,
		onchange,
		...props
	}: Props = $props();

	function handleChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		value = target.value;
		onchange?.(value);
	}
</script>

<select
	bind:value
	{disabled}
	{required}
	onchange={handleChange}
	class={cn(
		'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...props}
>
	{#if placeholder}
		<option value="" disabled>{placeholder}</option>
	{/if}
	{#each options as option}
		<option value={option.value}>{option.label}</option>
	{/each}
</select>
