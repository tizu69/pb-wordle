<script lang="ts">
	import '$lib/pocketbase';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>pb-wordle</title>
</svelte:head>

{@render children?.()}
