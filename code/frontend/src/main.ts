import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { setupGlobalErrorRuntime } from './runtime/globalErrorRuntime'
import { useAuthStore } from './stores/auth'
import './style.css'
import './assets/styles/theme.css'
import './assets/styles/surface-shell-background.css'
import './assets/styles/teacher-surface.css'
import './assets/styles/page-tabs.css'
import './assets/styles/workspace-shell.css'
import './assets/styles/workspace-glass.css'
import './assets/styles/journal-eyebrows.css'
import './assets/styles/journal-notes.css'
import './assets/styles/journal-soft-surfaces.css'
import './assets/styles/journal-admin-shell.css'
import './assets/styles/journal-user-shell.css'
import './assets/styles/journal-user-directory.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
setupGlobalErrorRuntime(app, router, pinia)

// Kick off silent session restore early. Router guards await the same restore promise when needed.
void useAuthStore(pinia).restore()

app.mount('#app')
