<script lang="ts">
	import Button from '$lib/components/ui/button.svelte';
	import Card from '$lib/components/ui/card.svelte';
	import { goto } from '$app/navigation';
	import { Database, Users, Zap, Code, CheckCircle } from 'lucide-svelte';
</script>

<svelte:head>
	<title>EzModel - Visual Database Schema Designer</title>
</svelte:head>

<!-- Hero Section -->
<section class="container mx-auto px-4 pt-16 pb-24">
	<div class="text-center max-w-4xl mx-auto">
		<h1 class="text-4xl md:text-6xl font-bold tracking-tight mb-6">
			Design Database Schemas
			<span class="text-primary">Visually</span>
		</h1>
		<p class="text-xl md:text-2xl text-muted-foreground mb-8 max-w-2xl mx-auto">
			EzModel is like Figma for databases. Create, collaborate, and generate SQL schemas with an intuitive visual interface.
		</p>

		<div class="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12">
			<Button size="lg" onclick={() => goto('/register')}>
				Get Started Free
			</Button>
			<Button variant="outline" size="lg" onclick={() => goto('/login')}>
				Sign In
			</Button>
		</div>

		<!-- Hero Image Placeholder -->
		<div class="rounded-lg border bg-muted/50 p-8 text-center text-muted-foreground">
			<Database class="mx-auto h-24 w-24 mb-4" />
			<p>Visual Schema Designer Preview</p>
		</div>
	</div>
</section>

<!-- Features Section -->
<section class="bg-muted/30 py-24">
	<div class="container mx-auto px-4">
		<div class="text-center mb-16">
			<h2 class="text-3xl md:text-4xl font-bold mb-4">
				Everything you need to design databases
			</h2>
			<p class="text-xl text-muted-foreground max-w-2xl mx-auto">
				From concept to production, EzModel provides all the tools you need for database schema design.
			</p>
		</div>

		<div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Database class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Visual Design</h3>
				<p class="text-muted-foreground">
					Drag and drop tables, define relationships, and design your database schema with an intuitive visual interface.
				</p>
			</Card>

			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Users class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Real-time Collaboration</h3>
				<p class="text-muted-foreground">
					Work together with your team in real-time. See live cursors, changes, and collaborate seamlessly.
				</p>
			</Card>

			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Code class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Multi-Database Support</h3>
				<p class="text-muted-foreground">
					Generate SQL for PostgreSQL, MySQL, SQLite, and SQL Server. Export schemas and migrations.
				</p>
			</Card>

			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Zap class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Code Generation</h3>
				<p class="text-muted-foreground">
					Automatically generate ORM models, API schemas, and database migrations from your visual designs.
				</p>
			</Card>

			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<CheckCircle class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Schema Validation</h3>
				<p class="text-muted-foreground">
					Built-in validation ensures your schemas are consistent and follow database best practices.
				</p>
			</Card>

			<Card class="p-6">
				<div class="mb-4">
					<div class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
						<Database class="h-6 w-6 text-primary" />
					</div>
				</div>
				<h3 class="text-xl font-semibold mb-2">Version Control</h3>
				<p class="text-muted-foreground">
					Track schema changes over time, create branches, and manage database evolution with confidence.
				</p>
			</Card>
		</div>
	</div>
</section>

<!-- CTA Section -->
<section class="py-24">
	<div class="container mx-auto px-4 text-center">
		<h2 class="text-3xl md:text-4xl font-bold mb-4">
			Ready to start designing?
		</h2>
		<p class="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
			Join thousands of developers who use EzModel to design better databases faster.
		</p>

		<Button size="lg" onclick={() => goto('/register')}>
			<Database class="mr-2 h-4 w-4" />
			Get Started Free
		</Button>
	</div>
</section>
