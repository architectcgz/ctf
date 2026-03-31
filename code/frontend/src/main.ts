import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import { redirectToErrorStatusPage } from './utils/errorStatusPage'
import './style.css'
import './assets/styles/theme.css'
import './assets/styles/element-override.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// 全局错误处理
app.config.errorHandler = (err, instance, info) => {
  console.error('Vue error:', err, info)
  redirectToErrorStatusPage(500)
}

// Restore the access token early; user profile is loaded lazily by router guards.
useAuthStore(pinia).restore()

app.mount('#app')
