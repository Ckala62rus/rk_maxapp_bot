// App entrypoint for Vue 3.
import { createApp } from "vue";
import App from "./App.vue";
import "./assets/styles.css";

// Mount root component.
const app = createApp(App);
app.mount("#app");

// Notify host webview (MAX/Telegram) that the app is ready.
if (window?.WebApp?.ready) {
  window.WebApp.ready();
}
if (window?.Telegram?.WebApp?.ready) {
  window.Telegram.WebApp.ready();
}
