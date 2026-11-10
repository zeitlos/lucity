<script setup lang="ts">
const log = [
  { time: '12:04:31', source: 'build', text: 'detected a Next.js app' },
  { time: '12:04:31', source: 'secrets', text: 'first scan, reading 214 commits' },
  { time: '12:04:58', source: 'secrets', text: 'AWS access key in config/settings.py:42', tone: 'c-warm' },
  { time: '12:04:59', source: 'secrets', text: 'asked AWS whether it still works: yes', tone: 'c-warm' },
  { time: '12:05:02', source: 'secrets', text: 'done, 1 key that works, 2 to review' },
  { time: '12:05:43', source: 'build', text: 'image ready in 1m 12s' },
  { time: '12:05:44', source: 'deploy', text: 'rolling out 2 instances' },
  { time: '12:06:25', source: 'deploy', text: 'live in 41s', tone: 'c-ok' },
];

const findings = [
  { name: 'AWS access key', masked: 'AKIA••••••••••••7QXF', where: 'config/settings.py:42', commit: '3f2a1c9', who: 'a.keller · 14 months ago', state: 'Still works', live: true },
  { name: 'Stripe secret key', masked: 'sk_live_••••••••••••4eC3', where: 'scripts/refund.rb:7', commit: 'b81c02d', who: 'm.frei · 5 months ago', state: 'No longer works', live: false },
  { name: 'Private key', masked: '-----BEGIN RSA••••••••', where: 'deploy/id_rsa:1', commit: '77e0a4b', who: 'a.keller · 2 years ago', state: "Can't be checked", live: false },
];

const features = [
  { name: 'The whole history, then what is new', text: 'The first deploy reads every commit. After that, only the commits since the last scan.' },
  { name: 'Exactly where', text: 'File, line, commit and author for every finding, with the value masked.' },
  { name: 'Keys that work, first', text: "A read-only question to the key's own provider separates live keys from old test fixtures." },
  { name: 'Never in the way', text: 'The release ships either way. You decide what to fix, and when.' },
];
</script>

<template>
  <section id="keys" class="mx-auto max-w-content scroll-mt-28 px-6" aria-labelledby="keys-title">
    <SecurityChapterHeading id="keys-title" beat="grow" promise="Leaked keys" rest="caught on every deploy.">
      Each deploy reads your repository for passwords, API keys and tokens, next to the build. What it finds lands on the Security tab of the service, with the keys that still work at the top.
    </SecurityChapterHeading>

    <LandingStage tint="grow" class="mt-12 px-4 pt-6 pb-10 sm:px-9 sm:pt-9">
      <div class="keys">
        <SecurityTerminal
          class="keys-term"
          title="Deployment 142 · main · 9b1e4d0"
          role="img"
          aria-label="The log of deployment 142: the build detects a Next.js app, the secret scan reads 214 commits, finds an AWS access key in config/settings.py line 42 and confirms with AWS that it still works, the image is ready after 1 minute 12, and the deploy is live 41 seconds later."
        >
          <div v-for="line in log" :key="line.time + line.text" class="log-line">
            <span class="c-dim log-time">{{ line.time }}</span>
            <span class="log-source">{{ line.source }}</span>
            <span :class="line.tone">{{ line.text }}</span>
          </div>
        </SecurityTerminal>

        <div class="keys-note">
          <h3 class="font-display text-2xl/tight text-stone-900">Rotate first.</h3>
          <p class="mt-2.5 text-base text-[oklch(0.38_0.05_60)]">
            Deleting the file does not take a key out of your history. Rotating it is what makes it harmless, and the finding stays for as long as the commit does.
          </p>
        </div>

        <div class="keys-tab" aria-hidden="true">
          <div class="tab-top"><span>Leaked secrets</span><small>web · production · scanned with deployment 142</small></div>
          <div v-for="finding in findings" :key="finding.name" class="finding">
            <span class="finding-icon"><UIcon name="i-lucide-key-round" class="size-4.5" /></span>
            <div class="min-w-0">
              <b>{{ finding.name }}</b>
              <span class="finding-mask">{{ finding.masked }}</span>
              <small><span class="font-mono">{{ finding.where }}</span> · <span class="font-mono">{{ finding.commit }}</span> · {{ finding.who }}</small>
            </div>
            <em class="badge" :class="finding.live ? 'badge-crit' : 'badge-off'">{{ finding.state }}</em>
          </div>
        </div>
      </div>
    </LandingStage>

    <SecurityFeatures :items="features" class="mt-8" />
  </section>
</template>

<style scoped>
.keys {
  position: relative;
  height: 460px;
}

.keys-term {
  position: absolute;
  left: 0;
  top: 0;
  z-index: 2;
  width: 47%;
}

.log-line {
  display: grid;
  grid-template-columns: 8ch 7ch minmax(0, 1fr);
  gap: 0 2ch;
  white-space: nowrap;
}

.keys-note {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 38%;
}

.keys-tab {
  position: absolute;
  left: 43%;
  right: -60px;
  top: 84px;
  bottom: -40px;
  overflow: hidden;
  border-radius: 18px 0 0;
  background: oklch(0.985 0.008 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.tab-top {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 2px 12px;
  padding: 20px 96px 16px 76px;
  border-bottom: 1px solid oklch(0.91 0.015 80);
  font-size: 17px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.tab-top small {
  font-size: 14.5px;
  font-weight: 400;
  color: oklch(0.47 0.04 55);
}

.finding {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr) auto;
  gap: 0 16px;
  align-items: start;
  padding: 18px 96px 18px 76px;
  border-top: 1px solid oklch(0.91 0.015 80);
}

.tab-top + .finding {
  border-top: 0;
}

.finding-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: oklch(0.93 0.07 85);
  color: oklch(0.42 0.10 70);
}

.finding b {
  display: block;
  font-size: 16.5px;
  font-weight: 600;
  line-height: 1.55;
  color: oklch(0.25 0.03 55);
}

.finding-mask {
  display: inline-block;
  margin-top: 5px;
  padding: 2px 8px;
  border-radius: 5px;
  background: oklch(0.95 0.012 80);
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 13px;
  line-height: 1.55;
  color: oklch(0.35 0.03 55);
  white-space: nowrap;
}

.finding small {
  display: block;
  margin-top: 6px;
  font-size: 14px;
  line-height: 1.55;
  color: oklch(0.47 0.04 55);
  white-space: nowrap;
}

.finding small .font-mono {
  font-size: 13px;
  color: oklch(0.35 0.03 55);
}

.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 13px;
  font-style: normal;
  font-weight: 600;
  line-height: 1.55;
  white-space: nowrap;
}

.badge-crit {
  background: oklch(0.55 0.19 27 / .12);
  color: oklch(0.47 0.19 27);
}

.badge-off {
  background: oklch(0.91 0.015 75);
  color: oklch(0.40 0.03 55);
}

@media (max-width: 1239px) {
  .keys {
    display: grid;
    grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr);
    gap: 28px 40px;
    align-items: end;
    height: auto;
  }

  .keys-term,
  .keys-note,
  .keys-tab {
    position: static;
    width: auto;
  }

  .keys-note {
    padding-bottom: 6px;
  }

  .keys-tab {
    grid-column: 1 / -1;
    margin: 0 0 -40px;
    padding-bottom: 40px;
    border-radius: 18px 18px 0 0;
  }

  .tab-top,
  .finding {
    padding-right: 24px;
    padding-left: 24px;
  }
}

@media (max-width: 767px) {
  .keys {
    grid-template-columns: minmax(0, 1fr);
    gap: 24px;
  }

  .keys-note {
    padding-bottom: 0;
  }

  .tab-top,
  .finding {
    padding-right: 20px;
    padding-left: 20px;
  }

  .finding small,
  .finding-mask {
    white-space: normal;
  }
}

@media (max-width: 560px) {
  .log-line {
    grid-template-columns: 7ch minmax(0, 1fr);
    gap: 0 1.5ch;
    white-space: normal;
  }

  .log-time {
    display: none;
  }

  .finding {
    grid-template-columns: 40px minmax(0, 1fr);
  }

  .finding .badge {
    grid-column: 2;
    justify-self: start;
    margin-top: 10px;
  }

  .finding-mask {
    overflow-wrap: anywhere;
  }
}
</style>
