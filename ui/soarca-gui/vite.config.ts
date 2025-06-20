import { defineConfig, loadEnv, type UserConfig, type ConfigEnv } from 'vite';
import react from '@vitejs/plugin-react-swc';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig(({ mode }: ConfigEnv): UserConfig => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [
      react(),
      tailwindcss(),
    ],
    define: {
      'process.env.SOARCA_URI': JSON.stringify(env.SOARCA_URI),
    },
    server: {
      port: 5173,
      host: true,
      watch: {
        usePolling: true,
      },
    },
    esbuild: {
      target: 'esnext',
    }
  };
});
