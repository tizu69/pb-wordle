<script lang="ts" module>
	export let toast = (msg: string) => {};
</script>

<script lang="ts">
	let msgs = $state<string[]>([]);
	toast = (msg: string) => msgs.push(msg);

	$effect(() => {
		if (msgs.length > 0)
			setTimeout(() => {
				msgs.shift();
			}, 3000);
	});
</script>

{#if msgs.length > 0}
	<div
		class={[
			'fixed inset-x-0 top-4 mx-auto w-fit border-4 bg-background',
			'z-50 rounded-lg p-2'
		]}
	>
		<p class="text-sm">{msgs[0]}</p>
	</div>
{/if}
