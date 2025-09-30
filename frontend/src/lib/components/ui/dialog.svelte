<script lang="ts">
	import { cn } from '$lib/utils/cn';
	import Button from './button.svelte';
	import { X } from 'lucide-svelte';

	type Props = {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		class?: string;
		children: any;
	};

	let {
		open = $bindable(false),
		onOpenChange,
		class: className,
		children,
		...props
	}: Props = $props();

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			close();
		}
	}

	function close() {
		open = false;
		onOpenChange?.(false);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			close();
		}
	}
</script>

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="dialog"
		tabindex="-1"
	>
		<!-- Dialog Content -->
		<div class="fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 p-4">
			<div
				class={cn(
					'relative grid w-full gap-4 rounded-lg border bg-background p-6 shadow-lg',
					className
				)}
				{...props}
			>
				<Button variant="ghost" size="icon" class="absolute right-4 top-4 h-6 w-6" onclick={close}>
					<X class="h-4 w-4" />
					<span class="sr-only">Close</span>
				</Button>
				{@render children()}
			</div>
		</div>
	</div>
{/if}
