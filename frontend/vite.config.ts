import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	const target = loadEnv(mode, '.', 'API_PROXY_TARGET').API_PROXY_TARGET ?? 'http://127.0.0.1:8989';
	return {
	server: {
		proxy: { '/api': { target, changeOrigin: true, rewrite: (path) => path.replace(/^\/api(?=\/|$)/, '') } }
	},
	preview: {
		proxy: { '/api': { target, changeOrigin: true, rewrite: (path) => path.replace(/^\/api(?=\/|$)/, '') } }
	},
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			adapter: adapter()
		})
	]
	};
});
