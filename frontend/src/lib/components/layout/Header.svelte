<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { authService } from '$lib/services/auth';
	import Button from '../ui/button.svelte';
	import { goto } from '$app/navigation';
	import { Database, LogOut, User } from 'lucide-svelte';

	// Remove props, use store directly

	function handleLogout() {
		authService.logout();
		authStore.clear();
		goto('/');
	}
</script>

<header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
	<div class="container mx-auto flex h-16 max-w-screen-2xl items-center justify-between px-4">
		<div class="flex items-center gap-2">
			<a href="/" class="flex items-center gap-2 font-bold text-xl">
				<Database class="h-6 w-6" />
				EzModel
			</a>
		</div>

		<nav class="flex items-center gap-4">
			{#if $authStore.isAuthenticated}
				<a href="/dashboard" class="text-sm font-medium hover:text-primary">
					Dashboard
				</a>
				<div class="flex items-center gap-2">
					<span class="text-sm text-muted-foreground">
						{$authStore.user?.username}
					</span>
					<Button variant="ghost" size="icon" onclick={() => goto('/profile')}>
						<User class="h-4 w-4" />
					</Button>
					<Button variant="ghost" size="icon" onclick={handleLogout}>
						<LogOut class="h-4 w-4" />
					</Button>
				</div>
			{:else}
				<a href="/login" class="text-sm font-medium hover:text-primary">
					Login
				</a>
				<Button onclick={() => goto('/register')}>
					Sign Up
				</Button>
			{/if}
		</nav>
	</div>
</header>