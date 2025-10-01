<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { authStore } from '$lib/stores/auth';
	import Header from '$lib/components/layout/Header.svelte';
	import Toast from '$lib/components/ui/toast.svelte';

	let { children } = $props();

	onMount(() => {
		authStore.init();
	});
</script>

<svelte:head>
	<title>EzModel - Visual Database Schema Designer</title>
	<meta
		name="description"
		content="Design database schemas visually with real-time collaboration"
	/>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="min-h-screen bg-background">
	{#if !$page.url.pathname.includes('/edit')}
		<Header />
	{/if}
	<main>
		{@render children?.()}
	</main>
	<Toast />
</div>
