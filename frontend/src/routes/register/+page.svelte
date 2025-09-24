<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { uiStore } from '$lib/stores/ui';
	import { authService } from '$lib/services/auth';
	import Button from '$lib/components/ui/button.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import { goto } from '$app/navigation';
	import { UserPlus, Database } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import type { RegisterRequest } from '$lib/types/api';

	// Remove props, use store directly

	let email = '';
	let username = '';
	let password = '';
	let confirmPassword = '';
	let isLoading = false;
	let emailError = '';
	let usernameError = '';
	let passwordError = '';
	let confirmPasswordError = '';

	// Redirect if already authenticated
	onMount(() => {
		if ($authStore.isAuthenticated) {
			goto('/dashboard');
		}
	});

	function validateForm() {
		emailError = '';
		usernameError = '';
		passwordError = '';
		confirmPasswordError = '';

		if (!email) {
			emailError = 'Email is required';
		} else if (!/\S+@\S+\.\S+/.test(email)) {
			emailError = 'Please enter a valid email address';
		}

		if (!username) {
			usernameError = 'Username is required';
		} else if (username.length < 3) {
			usernameError = 'Username must be at least 3 characters';
		} else if (!/^[a-zA-Z0-9_]+$/.test(username)) {
			usernameError = 'Username can only contain letters, numbers, and underscores';
		}

		if (!password) {
			passwordError = 'Password is required';
		} else if (password.length < 6) {
			passwordError = 'Password must be at least 6 characters';
		}

		if (!confirmPassword) {
			confirmPasswordError = 'Please confirm your password';
		} else if (password !== confirmPassword) {
			confirmPasswordError = 'Passwords do not match';
		}

		return !emailError && !usernameError && !passwordError && !confirmPasswordError;
	}

	async function handleRegister() {
		if (!validateForm()) return;

		isLoading = true;
		try {
			const registerData: RegisterRequest = { email, username, password };
			await authService.register(registerData);

			uiStore.success('Account Created!', 'Please sign in to continue');
			goto('/login');
		} catch (error: any) {
			uiStore.error('Registration Failed', error.message || 'Unable to create account');
			console.error('Registration error:', error);
		} finally {
			isLoading = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			handleRegister();
		}
	}
</script>

<svelte:head>
	<title>Sign Up - EzModel</title>
</svelte:head>

<div class="container mx-auto px-4 py-16">
	<div class="max-w-md mx-auto">
		<div class="text-center mb-8">
			<div class="flex items-center justify-center gap-2 mb-4">
				<Database class="h-8 w-8" />
				<span class="text-2xl font-bold">EzModel</span>
			</div>
			<h1 class="text-2xl font-bold mb-2">Create your account</h1>
			<p class="text-muted-foreground">Get started with EzModel for free</p>
		</div>

		<Card class="p-6">
			<form on:submit|preventDefault={handleRegister} class="space-y-4">
				<div class="space-y-2">
					<label for="email" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
						Email
					</label>
					<Input
						id="email"
						type="email"
						placeholder="Enter your email"
						bind:value={email}
						required
						disabled={isLoading}
						onkeydown={handleKeydown}
						class={emailError ? 'border-destructive' : ''}
					/>
					{#if emailError}
						<p class="text-sm text-destructive">{emailError}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<label for="username" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
						Username
					</label>
					<Input
						id="username"
						type="text"
						placeholder="Choose a username"
						bind:value={username}
						required
						disabled={isLoading}
						onkeydown={handleKeydown}
						class={usernameError ? 'border-destructive' : ''}
					/>
					{#if usernameError}
						<p class="text-sm text-destructive">{usernameError}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<label for="password" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
						Password
					</label>
					<Input
						id="password"
						type="password"
						placeholder="Create a password"
						bind:value={password}
						required
						disabled={isLoading}
						onkeydown={handleKeydown}
						class={passwordError ? 'border-destructive' : ''}
					/>
					{#if passwordError}
						<p class="text-sm text-destructive">{passwordError}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<label for="confirmPassword" class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
						Confirm Password
					</label>
					<Input
						id="confirmPassword"
						type="password"
						placeholder="Confirm your password"
						bind:value={confirmPassword}
						required
						disabled={isLoading}
						onkeydown={handleKeydown}
						class={confirmPasswordError ? 'border-destructive' : ''}
					/>
					{#if confirmPasswordError}
						<p class="text-sm text-destructive">{confirmPasswordError}</p>
					{/if}
				</div>

				<Button
					type="submit"
					class="w-full"
					disabled={isLoading}
				>
					{#if isLoading}
						<div class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent mr-2"></div>
						Creating account...
					{:else}
						<UserPlus class="mr-2 h-4 w-4" />
						Create Account
					{/if}
				</Button>
			</form>

			<div class="mt-6 text-center text-sm">
				<span class="text-muted-foreground">Already have an account? </span>
				<a href="/login" class="font-medium text-primary hover:underline">
					Sign in
				</a>
			</div>
		</Card>
	</div>
</div>