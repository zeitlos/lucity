<script setup lang="ts">
type Kind = 'service' | 'database' | 'bucket';

const icons: Record<Kind, string> = {
  service: 'i-simple-icons-github',
  database: 'i-lucide-database',
  bucket: 'i-lucide-archive',
};

const workspaces: { name: string; environments: { name: string; parts: { kind: Kind; name: string }[] }[] }[] = [
  {
    name: 'Workspace A',
    environments: [
      { name: 'development', parts: [{ kind: 'service', name: 'web' }, { kind: 'service', name: 'worker' }, { kind: 'database', name: 'database' }] },
      { name: 'production', parts: [{ kind: 'service', name: 'web' }, { kind: 'service', name: 'worker' }, { kind: 'database', name: 'database' }] },
    ],
  },
  {
    name: 'Workspace B',
    environments: [
      { name: 'staging', parts: [{ kind: 'service', name: 'api' }, { kind: 'database', name: 'database' }] },
      { name: 'production', parts: [{ kind: 'service', name: 'api' }, { kind: 'database', name: 'database' }, { kind: 'bucket', name: 'bucket' }] },
    ],
  },
];

const features = [
  { name: 'Workspaces are walled off', text: 'Apps in one cannot reach the apps or databases of another.' },
  { name: 'Environments are kept apart', text: 'The private network ends at the environment, so staging cannot reach production over it.' },
  { name: 'No special privileges', text: "Root inside your app is not root on the machine, and apps get no access to the platform's controls." },
  { name: 'Private databases', text: 'Reachable from your own services only, unless you open one to the addresses you list.' },
];

const mark = 'M 130 0 L 130 76 C 130 102.64, 114.4 112, 70 112 L 60 112 C 15.6 112, 0 121.36, 0 148 L 0 181 L 169 181 L 169 148 C 169 121.36, 178.23 112, 204.5 112 L 214.5 112 C 240.77 112, 250 102.64, 250 76 L 250 0 Z';
</script>

<template>
  <section id="walls" class="mx-auto max-w-content scroll-mt-28 px-6" aria-labelledby="walls-title">
    <SecurityChapterHeading id="walls-title" beat="joy" promise="Walls" rest="between everything.">
      Lucity runs code from many customers on shared machines, so isolation is the first design rule and not a setting you find later.
    </SecurityChapterHeading>

    <LandingStage tint="joy" class="mt-12 px-4 py-6 sm:px-10 sm:pt-10 sm:pb-11">
      <div
        class="iso"
        role="img"
        aria-label="The Lucity platform runs on its own machines, walled off from customer workspaces. Workspace A and workspace B are walled off from each other. Inside a workspace, every environment has its own services, databases, variables and private network."
      >
        <div class="platform">
          <b><svg width="18" height="18" viewBox="0 0 250 181" class="shrink-0"><path :d="mark" fill="#8843db" /></svg>The Lucity platform</b>
          <span>Runs on its own machines. Customer apps are never placed on them and cannot reach them.</span>
        </div>

        <template v-for="workspace in workspaces" :key="workspace.name">
          <div class="wall"><i><UIcon name="i-lucide-lock" class="size-4" /></i></div>
          <div class="workspace">
            <b>{{ workspace.name }}</b>
            <div class="environments">
              <div v-for="environment in workspace.environments" :key="environment.name" class="environment">
                <b>{{ environment.name }}</b>
                <span v-for="part in environment.parts" :key="part.name"><UIcon :name="icons[part.kind]" class="size-[15px] shrink-0" />{{ part.name }}</span>
                <small>own variables, own private network</small>
              </div>
            </div>
          </div>
        </template>
      </div>
    </LandingStage>

    <SecurityFeatures :items="features" class="mt-8" />

    <p class="mt-8 max-w-[980px] text-base text-stone-700 dark:text-stone-400">
      And when something goes wrong anyway: a database
      <NuxtLink to="/docs/postgres#backups" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">restores to any second of the last 30 days</NuxtLink>,
      a release
      <NuxtLink to="/docs/deployments" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">rolls back</NuxtLink>
      to the one before, and a broken deploy never replaces a working one.
    </p>
  </section>
</template>

<style scoped>
.iso {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 20px;
  line-height: 1.55;
}

.platform {
  flex: 1 1 220px;
  box-sizing: border-box;
  padding: 22px 24px;
  border-radius: 16px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.platform b {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 17px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.platform span {
  display: block;
  margin-top: 10px;
  font-size: 15px;
  line-height: 1.5;
  color: oklch(0.45 0.04 55);
}

.wall {
  position: relative;
  flex: none;
  width: 4px;
  border-radius: 2px;
  background: #8843db;
}

.wall i {
  position: absolute;
  left: 50%;
  top: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  margin: -17px 0 0 -17px;
  border-radius: 10px;
  background: #8843db;
  color: white;
}

.workspace {
  flex: 2 1 400px;
  box-sizing: border-box;
  padding: 20px 22px 22px;
  border-radius: 16px;
  background: oklch(0.99 0.006 80 / .55);
  box-shadow: inset 0 0 0 1.5px oklch(0.25 0.04 55 / .28);
}

.workspace > b {
  display: block;
  font-size: 17px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.dark .workspace {
  background: oklch(1 0 0 / .08);
  box-shadow: inset 0 0 0 1.5px oklch(1 0 0 / .14);
}

.dark .workspace > b {
  color: oklch(0.95 0.02 80);
}

.environments {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(150px, 100%), 1fr));
  gap: 14px;
  margin-top: 14px;
}

.environment {
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: 12px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12);
}

.environment b {
  display: block;
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.environment span {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
  font-size: 15px;
  color: oklch(0.35 0.03 55);
}

.environment small {
  display: block;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed oklch(0.55 0.15 300 / .5);
  font-size: 13.5px;
  line-height: 1.4;
  color: oklch(0.42 0.12 300);
}

@media (max-width: 900px) {
  .wall {
    width: 100%;
    height: 4px;
    margin: 13px 0;
  }

  .platform,
  .workspace {
    flex-basis: 100%;
  }
}

@media (max-width: 520px) {
  .workspace {
    padding: 16px 14px 14px;
  }

  .environments {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .environment {
    padding: 14px 12px;
  }
}
</style>
