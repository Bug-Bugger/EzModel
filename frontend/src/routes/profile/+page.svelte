<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { uiStore } from '$lib/stores/ui';
	import { apiClient } from '$lib/services/api';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import Input from '$lib/components/ui/input.svelte';
	import { User as UserIcon, Save, Key, Mail, UserCheck } from 'lucide-svelte';
	import type { UpdateUserRequest, UpdatePasswordRequest } from '$lib/types/api';
	import type { User } from '$lib/types/models';

	// Remove props, use store directly

	// Profile form
	let email = '';
	let username = '';
	let isUpdatingProfile = false;

	// Password form
	let currentPassword = '';
	let newPassword = '';
	let confirmNewPassword = '';
	let isUpdatingPassword = false;

	// Validation errors
	let emailError = '';
	let usernameError = '';
	let passwordErrors = {
		current: '',
		new: '',
		confirm: ''
	};

	// Redirect if not authenticated
	onMount(() => {
		if (!$authStore.isAuthenticated) {
			goto('/login');
			return;
		}

		if ($authStore.user) {
			email = $authStore.user.email;
			username = $authStore.user.username;
		}
	});

	function validateProfileForm() {
		emailError = '';
		usernameError = '';

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

		return !emailError && !usernameError;
	}

	function validatePasswordForm() {
		passwordErrors = {
			current: '',
			new: '',
			confirm: ''
		};

		if (!currentPassword) {
			passwordErrors.current = 'Current password is required';
		}

		if (!newPassword) {
			passwordErrors.new = 'New password is required';
		} else if (newPassword.length < 6) {
			passwordErrors.new = 'Password must be at least 6 characters';
		}

		if (!confirmNewPassword) {
			passwordErrors.confirm = 'Please confirm your new password';
		} else if (newPassword !== confirmNewPassword) {
			passwordErrors.confirm = 'Passwords do not match';
		}

		return !passwordErrors.current && !passwordErrors.new && !passwordErrors.confirm;
	}

	async function updateProfile() {
		if (!validateProfileForm()) return;

		isUpdatingProfile = true;
		try {
			const updateData: UpdateUserRequest = {
				email: email !== $authStore.user?.email ? email : undefined,
				username: username !== $authStore.user?.username ? username : undefined
			};

			// Only send request if there are changes
			if (updateData.email || updateData.username) {
				const response = await apiClient.put<User>(`/users/${$authStore.user?.id}`, updateData);

				if (response.success && response.data) {
					authStore.setUser(response.data);
					uiStore.success('Profile updated successfully!');
				}
			} else {
				uiStore.info('No changes to save');
			}
		} catch (error: any) {
			uiStore.error('Failed to update profile', error.message);
		} finally {
			isUpdatingProfile = false;
		}
	}

	async function updatePassword() {
		if (!validatePasswordForm()) return;

		isUpdatingPassword = true;
		try {
			const passwordData: UpdatePasswordRequest = {
				current_password: currentPassword,
				new_password: newPassword
			};

			await apiClient.put(`/users/${$authStore.user?.id}/password`, passwordData);

			// Clear form
			currentPassword = '';
			newPassword = '';
			confirmNewPassword = '';

			uiStore.success('Password updated successfully!');
		} catch (error: any) {
			uiStore.error('Failed to update password', error.message);
		} finally {
			isUpdatingPassword = false;
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<svelte:head>
	<title>Profile - EzModel</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 max-w-2xl">
	<!-- Header -->
	<div class="mb-8">
		<h1 class="text-3xl font-bold">Profile Settings</h1>
		<p class="text-muted-foreground mt-2">Manage your account settings and preferences</p>
	</div>

	<div class="space-y-6">
		<!-- Profile Information -->
		<Card class="p-6">
			<div class="flex items-center gap-3 mb-6">
				<div class="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
					<UserIcon class="h-5 w-5 text-primary" />
				</div>
				<div>
					<h2 class="text-xl font-semibold">Profile Information</h2>
					<p class="text-sm text-muted-foreground">Update your account details</p>
				</div>
			</div>

			<form on:submit|preventDefault={updateProfile} class="space-y-4">
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-2">
						<label for="email" class="text-sm font-medium">Email</label>
						<Input
							id="email"
							type="email"
							bind:value={email}
							disabled={isUpdatingProfile}
							class={emailError ? 'border-destructive' : ''}
						/>
						{#if emailError}
							<p class="text-sm text-destructive">{emailError}</p>
						{/if}
					</div>

					<div class="space-y-2">
						<label for="username" class="text-sm font-medium">Username</label>
						<Input
							id="username"
							type="text"
							bind:value={username}
							disabled={isUpdatingProfile}
							class={usernameError ? 'border-destructive' : ''}
						/>
						{#if usernameError}
							<p class="text-sm text-destructive">{usernameError}</p>
						{/if}
					</div>
				</div>

				<div class="flex justify-end">
					<Button type="submit" disabled={isUpdatingProfile}>
						{#if isUpdatingProfile}
							<div
								class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent mr-2"
							></div>
							Updating...
						{:else}
							<Save class="mr-2 h-4 w-4" />
							Update Profile
						{/if}
					</Button>
				</div>
			</form>
		</Card>

		<!-- Change Password -->
		<Card class="p-6">
			<div class="flex items-center gap-3 mb-6">
				<div class="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
					<Key class="h-5 w-5 text-primary" />
				</div>
				<div>
					<h2 class="text-xl font-semibold">Change Password</h2>
					<p class="text-sm text-muted-foreground">Update your account password</p>
				</div>
			</div>

			<form on:submit|preventDefault={updatePassword} class="space-y-4">
				<div class="space-y-2">
					<label for="currentPassword" class="text-sm font-medium">Current Password</label>
					<Input
						id="currentPassword"
						type="password"
						placeholder="Enter your current password"
						bind:value={currentPassword}
						disabled={isUpdatingPassword}
						class={passwordErrors.current ? 'border-destructive' : ''}
					/>
					{#if passwordErrors.current}
						<p class="text-sm text-destructive">{passwordErrors.current}</p>
					{/if}
				</div>

				<div class="grid gap-4 sm:grid-cols-2">
					<div class="space-y-2">
						<label for="newPassword" class="text-sm font-medium">New Password</label>
						<Input
							id="newPassword"
							type="password"
							placeholder="Enter new password"
							bind:value={newPassword}
							disabled={isUpdatingPassword}
							class={passwordErrors.new ? 'border-destructive' : ''}
						/>
						{#if passwordErrors.new}
							<p class="text-sm text-destructive">{passwordErrors.new}</p>
						{/if}
					</div>

					<div class="space-y-2">
						<label for="confirmNewPassword" class="text-sm font-medium">Confirm New Password</label>
						<Input
							id="confirmNewPassword"
							type="password"
							placeholder="Confirm new password"
							bind:value={confirmNewPassword}
							disabled={isUpdatingPassword}
							class={passwordErrors.confirm ? 'border-destructive' : ''}
						/>
						{#if passwordErrors.confirm}
							<p class="text-sm text-destructive">{passwordErrors.confirm}</p>
						{/if}
					</div>
				</div>

				<div class="flex justify-end">
					<Button type="submit" disabled={isUpdatingPassword}>
						{#if isUpdatingPassword}
							<div
								class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent mr-2"
							></div>
							Updating...
						{:else}
							<Key class="mr-2 h-4 w-4" />
							Update Password
						{/if}
					</Button>
				</div>
			</form>
		</Card>

		<!-- Account Information -->
		{#if $authStore.user}
			<Card class="p-6">
				<div class="flex items-center gap-3 mb-4">
					<div class="h-10 w-10 rounded-lg bg-primary/10 flex items-center justify-center">
						<UserCheck class="h-5 w-5 text-primary" />
					</div>
					<div>
						<h2 class="text-xl font-semibold">Account Information</h2>
						<p class="text-sm text-muted-foreground">Your account details</p>
					</div>
				</div>

				<div class="space-y-3">
					<div class="flex justify-between items-center py-2">
						<span class="text-sm font-medium">Account ID</span>
						<span class="text-sm text-muted-foreground font-mono">{$authStore.user.id}</span>
					</div>
				</div>
			</Card>
		{/if}
	</div>
</div>
