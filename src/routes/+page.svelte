<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import pb, { type CollCategory, type CollWord } from '$lib/pocketbase';
	import Toast, { toast } from '$lib/Toast.svelte';
	import { ClientResponseError } from 'pocketbase';

	let category = $state<CollCategory>();
	let word = $state<CollWord>();

	const categories = pb
		.collection('categories')
		.getFullList({ sort: '-created' });
	categories.then((cats) => {
		if (page.url.searchParams.has('c')) {
			const v = page.url.searchParams.get('c');
			category = cats.find((c) => c.name === v || c.id === v);
		}
		if (!category) category = cats[0];
	});

	let wordPromise = $state(Promise.resolve());
	function regenWord() {
		hintRequested = false;
		attempts = [];
		curAttempt = 0;
		wordPromise = pb
			.collection('words')
			.getFirstListItem(pb.filter(`category = {:c}`, { c: category?.id }), {
				sort: '@random'
			})
			.then((w) => {
				w.word = w.word.toUpperCase();
				word = w;
				toast('Next word!');
			});
	}
	$effect(() => {
		if (category) regenWord();
	});

	let hintRequested = $state(false);
	let attempts = $state<string[]>([]);
	let curAttempt = $state(0);
	let isFocused = $state(false);
	let showHint = $derived(hintRequested || category?.mustHint || false);
	let letters = $derived(word?.word.length || 0);
	let allowedAttempts = $derived(Math.ceil(letters * 1.2));
	let isDone = $derived(
		curAttempt >= allowedAttempts || word?.word === attempts[curAttempt - 1]
	);

	function escapeRegex(str: string) {
		return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
	}
	function removeRefs(str: string) {
		const safeWord = escapeRegex(word!.word);
		return str.replace(new RegExp(safeWord, 'gi'), '[...]');
	}
	function containsNonAscii(str: string) {
		return /[^ -~]/.test(str);
	}

	type Correctness = 0 | 1 | 2;
	function guessCorrectness(attempt: number): Correctness[] {
		const target = word?.word || '';
		const guess = attempts[attempt] || '';
		const targetCodes = [...target].map((c) => c.charCodeAt(0));
		const guessCodes = [...guess].map((c) => c.charCodeAt(0));
		const result = Array(targetCodes.length).fill(0);
		const used = Array(targetCodes.length).fill(false);
		guessCodes.forEach((code, i) => {
			if (targetCodes[i] === code) {
				result[i] = 2;
				used[i] = true;
			}
		});
		guessCodes.forEach((code, i) => {
			if (result[i]) return; // greren alr
			const foundIndex = targetCodes.findIndex(
				(targetCode, j) => targetCode === code && !used[j]
			);
			if (foundIndex !== -1) {
				result[i] = 1;
				used[foundIndex] = true;
			} else {
				result[i] = 0;
			}
		});
		return result;
	}
	function guessCorrectnessClasses(attempt: number): string[] {
		if (attempt >= curAttempt) return Array(letters).fill('border-2');
		return guessCorrectness(attempt).map((c) => {
			switch (c) {
				case 0:
					return 'bg-letter-nope';
				case 1:
					return 'bg-letter-partial';
				case 2:
					return 'bg-letter-done';
			}
		});
	}
</script>

<Toast />

<header class="flex items-center justify-between border-b-2 p-3">
	{#await categories}
		<p>Please wait...</p>
	{:then cats}
		<select
			class="rounded border-2 bg-background p-1"
			onchange={(e) =>
				(category = cats.find((c) => c.id == e.currentTarget.value))}
			value={category?.id}
		>
			{#each cats as cat}
				<option value={cat.id}>{cat.name}</option>
			{/each}
		</select>
	{/await}

	<div class="flex gap-2">
		{#if !showHint && word?.hint}
			<button onclick={() => (hintRequested = true)} class="underline">
				Hint
			</button>
		{/if}
		<button onclick={regenWord} class="underline">
			{isDone ? 'Next' : 'Skip'}
		</button>
	</div>
</header>

<main class="p-4 *:mx-auto *:w-fit">
	{#await wordPromise}
		<p class="pt-12">Please wait...</p>
	{:then}
		{#if word}
			{#if containsNonAscii(word.word) && !isDone}
				<div
					class={[
						'mb-4 grid grid-cols-[auto_1fr] gap-2 rounded-xl',
						'items-center border-2 p-1'
					]}
				>
					<p class="text-left font-bold">
						This word contains non-ASCII characters.
					</p>
					<button
						onclick={regenWord}
						class="rounded-lg bg-letter-partial px-2 py-1"
					>
						Skip
					</button>
				</div>
			{/if}

			{#if showHint}
				<p class="mb-1 text-lg font-bold">{removeRefs(word.hint || '')}</p>
				<p class="mb-4 max-w-prose text-justify text-sm">
					{removeRefs(word.hintLong || '')}
				</p>
			{/if}

			{#if !isFocused}
				<p class="text-xs font-bold">Click the grid to start typing!</p>
			{:else if isDone}
				<p class="text-xs">shift-enter for next!</p>
			{:else}
				<p class="text-xs">enter to submit, shift-enter to skip</p>
			{/if}

			<div
				class="relative mt-1 grid gap-1"
				style="grid-template-columns: repeat({letters}, minmax(0, 1fr))"
			>
				{#each { length: allowedAttempts } as _, attempt}
					{@const letterClass = guessCorrectnessClasses(attempt)}
					{#each { length: letters } as _, char}
						{@const letter = attempts?.[attempt]?.[char] || ''}
						<div
							class={[
								'grid size-12 place-items-center font-bold',
								letterClass[char]
							]}
						>
							{letter}
						</div>
					{/each}
				{/each}

				<input
					autofocus
					onkeypress={async (e) => {
						if (e.key != 'Enter') return;

						if (e.shiftKey) {
							e.currentTarget.value = '';
							return regenWord();
						}
						if (isDone) return;

						if (attempts[curAttempt]?.length != letters)
							return toast(`Not enough letters!`);

						const el = e.currentTarget;
						const attempt = attempts[curAttempt];
						if (category?.mustPresent)
							try {
								await pb.collection('words').getFirstListItem(
									pb.filter(`category = {:c} && word:lower = {:w}`, {
										c: category.id,
										w: attempt.toLowerCase()
									})
								);
							} catch (_err) {
								const err = _err as ClientResponseError;
								if (err?.status == 404) return toast('Not in word list!');
							}
						el.value = '';
						curAttempt++;

						if (attempt === word?.word) return toast('You won! Great job!! :3');
						if (curAttempt >= allowedAttempts)
							return toast(word?.word || '???');
					}}
					oninput={(e) => {
						e.preventDefault();
						if (isDone) return;
						attempts[curAttempt] = e.currentTarget.value
							.toUpperCase()
							.substring(0, letters);
						e.currentTarget.value = attempts[curAttempt];
					}}
					onfocus={() => (isFocused = true)}
					onblur={() => (isFocused = false)}
					type="text"
					class="absolute inset-0 z-10 h-full w-full opacity-0"
				/>
			</div>
		{/if}
	{/await}
</main>
