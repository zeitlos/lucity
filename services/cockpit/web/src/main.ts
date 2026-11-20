import './assets/main.css';

import { createApp } from 'vue';
import { DefaultApolloClient } from '@vue/apollo-composable';

import App from './App.vue';
import { apolloClient } from './lib/apollo';

createApp(App).provide(DefaultApolloClient, apolloClient).mount('#app');
