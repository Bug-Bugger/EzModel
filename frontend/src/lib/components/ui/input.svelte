<script lang="ts">
	import { cn } from '$lib/utils/cn';

	type Props = {
		type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url';
		placeholder?: string;
		value?: string;
		class?: string;
		disabled?: boolean;
		required?: boolean;
		id?: string;
		name?: string;
		readonly?: boolean;
		oninput?: (e: Event) => void;
	};

	let {
		type = 'text',
		placeholder,
		value = $bindable(''),
		class: className,
		disabled = false,
		required = false,
		id,
		name,
		readonly = false,
		oninput,
		...props
	}: Props = $props();
</script>

<input
	{type}
	{placeholder}
	bind:value
	{disabled}
	{required}
	{id}
	{name}
	{readonly}
	{oninput}
	class={cn(
		'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
	{...props}
/>
