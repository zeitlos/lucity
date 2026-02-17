import { ref, watch } from 'vue';

type Theme = 'light' | 'dark';

const STORAGE_KEY = 'lucity-theme';

function getInitialTheme(): Theme {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === 'light' || stored === 'dark') return stored;
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

const theme = ref<Theme>(getInitialTheme());

function applyTheme(t: Theme) {
  document.documentElement.classList.toggle('dark', t === 'dark');
}

applyTheme(theme.value);

watch(theme, (t) => {
  localStorage.setItem(STORAGE_KEY, t);
  applyTheme(t);
});

export function useTheme() {
  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light';
  }

  return { theme, toggleTheme };
}
