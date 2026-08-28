export default defineAppConfig({
  docus: {
    name: 'Lucity',
    description: 'Open-source PaaS on Kubernetes with full ejectability',
    url: 'https://lucity.cloud',
    socials: {
      github: 'zeitlos/lucity'
    }
  },
  seo: {
    titleTemplate: '%s | Lucity'
  },
  ui: {
    contentToc: {
      compoundVariants: [{
        color: 'primary',
        active: true,
        class: { link: 'text-highlighted font-medium' },
      }, {
        highlight: true,
        highlightVariant: 'straight',
        class: { indicator: 'w-0.5' },
      }],
    },
    colors: {
      primary: 'brand',
      neutral: 'stone',
      info: 'violet'
    },
    pageSection: {
      slots: {
        title: 'text-4xl sm:text-5xl lg:text-6xl text-pretty tracking-tight font-bold text-highlighted',
        description: 'text-base sm:text-lg text-muted text-center text-balance mt-6 mb-14'
      }
    },
    pageCard: {
      slots: {
        container: 'font-sans relative flex flex-col flex-1 lg:grid gap-x-10 gap-y-2 p-8 sm:p-12',
        wrapper: 'flex flex-col flex-1 items-start text-left',
        title: 'font-display text-3xl sm:text-4xl font-normal text-highlighted',
        description: 'font-sans text-base sm:text-lg text-pretty mt-4 !text-muted'
      }
    },
    pageHeader: {
      slots: {
        title: 'font-display font-normal text-4xl sm:text-5xl'
      }
    },
    prose: {
      img: {
        slots: {
          base: 'rounded-lg border border-default shadow-[0_2px_8px_oklch(0_0_0/0.05)]',
          zoomedImage: 'rounded-lg border border-default'
        }
      },
      h1: {
        slots: {
          base: 'font-display font-normal text-5xl'
        }
      },
      h2: {
        slots: {
          base: 'font-display font-normal text-3xl'
        }
      },
      h3: {
        slots: {
          base: 'font-display font-normal text-2xl'
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
