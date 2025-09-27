<script lang="ts">
	import { cn } from '$lib/utils/cn';

	type Props = {
		class?: string;
		children: any;
	};

	let { class: className, children, ...props }: Props = $props();
</script>

<div
	class={cn('rounded-xl border bg-card text-card-foreground shadow', className)}
	{...props}
>
	{@render children()}
</div>