import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			// in dev, vite takes preference, as such, this is used there.
			// in prod, pb serves the static files, so this will be ignored.
			'/api': { target: 'http://localhost:5555' },
			'/_': { target: 'http://localhost:5555' }
		}
	}
});
