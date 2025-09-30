<script lang="ts">
	import { uiStore } from '$lib/stores/ui';
	import { CheckCircle, XCircle, AlertTriangle, Info, X } from 'lucide-svelte';
	import { cn } from '$lib/utils/cn';
	import Button from './button.svelte';

	// Remove props, use store directly

	function getIcon(type: string) {
		switch (type) {
			case 'success':
				return CheckCircle;
			case 'error':
				return XCircle;
			case 'warning':
				return AlertTriangle;
			case 'info':
				return Info;
			default:
				return Info;
		}
	}

	function getToastClasses(type: string) {
		const baseClasses =
			'relative flex w-full items-center justify-between space-x-4 overflow-hidden rounded-md border p-4 pr-8 shadow-lg transition-all';

		switch (type) {
			case 'success':
				return cn(baseClasses, 'border-green-200 bg-green-50 text-green-900');
			case 'error':
				return cn(baseClasses, 'border-red-200 bg-red-50 text-red-900');
			case 'warning':
				return cn(baseClasses, 'border-yellow-200 bg-yellow-50 text-yellow-900');
			case 'info':
				return cn(baseClasses, 'border-blue-200 bg-blue-50 text-blue-900');
			default:
				return cn(baseClasses, 'border-gray-200 bg-gray-50 text-gray-900');
		}
	}
</script>

<div class="fixed top-0 right-0 z-50 w-full max-w-sm p-4">
	{#each $uiStore.toasts as toast (toast.id)}
		<div class={getToastClasses(toast.type)} data-toast-id={toast.id}>
			<div class="flex items-start gap-3">
				<svelte:component this={getIcon(toast.type)} class="mt-0.5 h-4 w-4 flex-shrink-0" />
				<div class="grid gap-1">
					<div class="text-sm font-semibold">{toast.title}</div>
					{#if toast.description}
						<div class="text-sm opacity-90">{toast.description}</div>
					{/if}
				</div>
			</div>
			<Button
				variant="ghost"
				size="icon"
				class="absolute right-2 top-2 h-6 w-6 rounded-full"
				onclick={() => uiStore.removeToast(toast.id)}
			>
				<X class="h-3 w-3" />
			</Button>
		</div>
	{/each}
</div>
