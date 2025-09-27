import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// Detect if running in Docker container
const isDocker = process.env.NODE_ENV === 'development' && process.env.DOCKER_ENV === 'true';

// Use appropriate backend URL based on environment
const backendUrl = isDocker ? 'http://backend:8080' : 'http://localhost:8080';

export default defineConfig({
	plugins: [sveltekit()],
	envDir: '../', // Load environment variables from project root
	server: {
		host: '0.0.0.0',
		port: 5173,
		proxy: {
			'/api': {
				target: backendUrl,
				changeOrigin: true,
				secure: false,
				ws: true // Enable WebSocket proxying
			}
		}
	}
});
