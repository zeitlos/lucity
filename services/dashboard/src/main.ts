import './assets/main.css';
import 'vue-sonner/style.css';

import { createApp } from 'vue';
import { DefaultApolloClient } from '@vue/apollo-composable';
import App from './App.vue';
import router from './router';
import { apolloClient } from './lib/apollo';

const app = createApp(App);

app.provide(DefaultApolloClient, apolloClient);
app.use(router);

app.mount('#app');
