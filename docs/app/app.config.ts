export default defineAppConfig({
  docus: {
    name: 'Lucity',
    description: 'Open-source PaaS on Kubernetes with full ejectability',
    url: 'https://lucity.dev',
    socials: {
      github: 'zeitlos/lucity'
    }
  },
  ui: {
    colors: {
      primary: 'teal',
      neutral: 'stone'
    },
    pageHeader: {
      slots: {
        title: 'text-4xl sm:text-5xl'
      }
    },
    prose: {
      h1: {
        slots: {
          base: 'text-5xl'
        }
      },
      h2: {
        slots: {
          base: 'text-3xl'
        }
      },
      h3: {
        slots: {
          base: 'text-2xl'
        }
      },
      h4: {
        slots: {
          base: 'text-xl'
        }
      }
    }
  },
  header: {
    title: 'Lucity',
    logo: {
      light: '/logo-light.svg',
      dark: '/logo-dark.svg',
      alt: 'Lucity'
    }
  }
});
