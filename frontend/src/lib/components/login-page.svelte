<script lang="ts">
	import AuthForm from './auth-form.svelte';
	let { busy, onauthenticate, error }: { busy: boolean; onauthenticate: (mode: 'login' | 'register', user: string, pass: string) => Promise<void>; error: string } = $props();
</script>

<main class="flex min-h-svh items-center justify-center bg-background p-4 sm:p-8">
	<div class="w-full max-w-4xl space-y-6">
		<h1 class="text-center text-2xl font-semibold tracking-tight">Портфель</h1>
		{#if error}<p role="alert" class="text-center text-sm text-destructive">{error}</p>{/if}
		<div class="grid items-start gap-6 md:grid-cols-2 md:gap-12"><AuthForm mode="login" {busy} {onauthenticate} /><AuthForm mode="register" {busy} {onauthenticate} /></div>
	</div>
</main>
