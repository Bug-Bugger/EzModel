<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { uiStore } from '$lib/stores/ui';
	import { authService } from '$lib/services/auth';
	import Button from '$lib/components/ui/button.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import { goto } from '$app/navigation';
	import { LogIn, Database } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import type { LoginRequest } from '$lib/types/api';

	// Remove props, use store directly

	let email = '';
	let password = '';
	let isLoading = false;
	let emailError = '';
	let passwordError = '';

	// Redirect if already authenticated
	onMount(() => {
		if ($authStore.isAuthenticated) {
			goto('/projects');
		}
	});

	function validateForm() {
		emailError = '';
		passwordError = '';

		if (!email) {
			emailError = 'Email is required';
		} else if (!/\S+@\S+\.\S+/.test(email)) {
			emailError = 'Please enter a valid email address';
		}

		if (!password) {
			passwordError = 'Password is required';
		} else if (password.length < 6) {
			passwordError = 'Password must be at least 6 characters';
		}

		return !emailError && !passwordError;
	}

	async function handleLogin() {
		if (!validateForm()) return;

		isLoading = true;
		try {
			const loginData: LoginRequest = { email, password };
			const user = await authService.login(loginData);

			authStore.setUser(user);
			uiStore.success('Welcome back!', `Logged in as ${user.username}`);
			goto('/projects');
		} catch (error: any) {
			uiStore.error('Login Failed', error.message || 'Invalid credentials');
			console.error('Login error:', error);
		} finally {
			isLoading = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleLogin();
		}
	}
</script>

<svelte:head>
	<title>Login - EzModel</title>
</svelte:head>

<div class="container mx-auto px-4 py-16">
	<div class="max-w-md mx-auto">
		<div class="text-center mb-8">
			<div class="flex items-center justify-center gap-2 mb-4">
				<Database class="h-8 w-8" />
				<span class="text-2xl font-bold">EzModel</span>
			</div>
			<h1 class="text-2xl font-bold mb-2">Welcome back</h1>
			<p class="text-muted-foreground">Sign in to your account to continue</p>
		</div>

		<Card class="p-6">
			<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
			<form on:submit|preventDefault={handleLogin} on:keydown={handleKeydown} class="space-y-4">
				<div class="space-y-2">
					<label
						for="email"
						class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
					>
						Email
					</label>
					<Input
						id="email"
						type="email"
						placeholder="Enter your email"
						bind:value={email}
						required
						disabled={isLoading}
						class={emailError ? 'border-destructive' : ''}
					/>
					{#if emailError}
						<p class="text-sm text-destructive">{emailError}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<label
						for="password"
						class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
					>
						Password
					</label>
					<Input
						id="password"
						type="password"
						placeholder="Enter your password"
						bind:value={password}
						required
						disabled={isLoading}
						class={passwordError ? 'border-destructive' : ''}
					/>
					{#if passwordError}
						<p class="text-sm text-destructive">{passwordError}</p>
					{/if}
				</div>

				<Button type="submit" class="w-full" disabled={isLoading}>
					{#if isLoading}
						<div
							class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent mr-2"
						></div>
						Signing in...
					{:else}
						<LogIn class="mr-2 h-4 w-4" />
						Sign In
					{/if}
				</Button>
			</form>

			<div class="mt-6 text-center text-sm">
				<span class="text-muted-foreground">Don't have an account? </span>
				<a href="/register" class="font-medium text-primary hover:underline"> Sign up </a>
			</div>
		</Card>
	</div>
</div>
