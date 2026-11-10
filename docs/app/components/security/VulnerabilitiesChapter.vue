<script setup lang="ts">
const counts = [
  { label: '1 critical', color: 'oklch(0.55 0.19 27)' },
  { label: '3 high', color: 'oklch(0.65 0.17 45)' },
  { label: '12 medium', color: 'oklch(0.78 0.14 85)' },
  { label: '20 low', color: 'oklch(0.70 0.10 250)' },
];

const findings = [
  { name: 'openssl', source: 'Base image', version: '3.0.13', fix: 'fixed in 3.0.15', severity: 'Critical' },
  { name: 'express', source: 'Dependency', version: '4.18.2', fix: 'fixed in 4.19.2', severity: 'High' },
  { name: 'jsonwebtoken', source: 'Dependency', version: '8.5.1', fix: 'fixed in 9.0.0', severity: 'High' },
  { name: 'libxml2', source: 'Base image', version: '2.9.14', fix: 'fixed in 2.9.14+deb12u1', severity: 'High' },
  { name: 'glibc', source: 'Base image', version: '2.36', fix: 'no fix yet', severity: 'Medium' },
  { name: 'semver', source: 'Dependency', version: '7.3.5', fix: 'fixed in 7.5.2', severity: 'Medium' },
  { name: 'tar', source: 'Dependency', version: '6.1.11', fix: 'fixed in 6.2.1', severity: 'Medium' },
];

const features = [
  { name: 'What sits under your code, too', text: 'Your dependencies and the base image below them, the system packages a look at your package files never sees.' },
  { name: 'The fix, not only the problem', text: 'Each finding names the version that fixes it, or says that there is none yet.' },
  { name: 'Worst first', text: 'Critical and high lead the list, each marked as base image or dependency.' },
  { name: 'Where you see it', text: 'On the Security tab of each service. Alerts by email are not built yet, and prebuilt images are not scanned yet.' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="vulnerabilities-title">
    <SecurityChapterHeading id="vulnerabilities-title" beat="keep" promise="Vulnerabilities" rest="checked while your app runs." :lead-width="500">
      Lucity rechecks every running app it built, every two hours. A vulnerability announced tomorrow shows up on the app you deployed last month.
    </SecurityChapterHeading>

    <LandingStage tint="keep" class="mt-12 px-4 pt-6 pb-10 sm:px-9 sm:pt-9">
      <div class="vul">
        <div class="vul-note">
          <h3 class="font-display text-2xl/tight text-stone-900">It keeps looking after you ship.</h3>
          <p class="mt-2.5 text-base text-[oklch(0.38_0.05_8)]">
            A check at build time goes quiet the moment the build is done. This one runs against what is live, so the list changes when the world does.
          </p>
        </div>

        <div class="vul-month">
          <b>web, deployed once, four weeks ago</b>
          <div class="bar" role="img" aria-label="Four weeks of rechecks every two hours. The app was deployed once at the start. A vulnerability announced in the third week appears on the Security tab at the next recheck.">
            <i class="mark-deployed" /><i class="mark-announced" />
          </div>
          <div class="bar-axis">
            <span class="at-deployed">deployed</span>
            <span class="at-announced">openssl fix announced</span>
            <span class="at-today">today</span>
          </div>
          <p>Rechecked every two hours since. No new deploy was needed for the finding to show up.</p>
        </div>

        <div class="vul-tab" aria-hidden="true">
          <div class="tab-top"><span>Vulnerabilities</span><small>web · production · checked 40 minutes ago</small></div>
          <div class="summary">
            <span v-for="count in counts" :key="count.label"><i :style="{ background: count.color }" />{{ count.label }}</span>
          </div>
          <div v-for="finding in findings" :key="finding.name" class="row">
            <b>{{ finding.name }}</b>
            <span class="row-source">
              <UIcon :name="finding.source === 'Base image' ? 'i-lucide-layers' : 'i-lucide-package'" class="size-[15px] shrink-0" />
              {{ finding.source }}
            </span>
            <span class="row-version">{{ finding.version }}, {{ finding.fix }}</span>
            <em class="badge" :class="`badge-${finding.severity.toLowerCase()}`">{{ finding.severity }}</em>
          </div>
        </div>
      </div>
    </LandingStage>

    <SecurityFeatures :items="features" class="mt-8" />
  </section>
</template>

<style scoped>
.vul {
  position: relative;
  height: 462px;
}

.vul-note {
  position: absolute;
  left: 0;
  top: 0;
  width: 32%;
}

.vul-month {
  position: absolute;
  left: 0;
  bottom: 0;
  z-index: 2;
  box-sizing: border-box;
  width: 40%;
  padding: 20px 24px 18px;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.vul-month b {
  display: block;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.55;
  color: oklch(0.25 0.03 55);
}

.vul-month p {
  margin-top: 12px;
  font-size: 14.5px;
  line-height: 1.45;
  color: oklch(0.45 0.04 55);
}

.bar {
  position: relative;
  height: 38px;
  margin-top: 14px;
  background-color: oklch(0.25 0.04 55 / .035);
  background-image: repeating-linear-gradient(to right, oklch(0.25 0.04 55 / .28) 0 1px, transparent 1px 6px);
  box-shadow: 0 1px 0 oklch(0.25 0.03 55);
}

.bar i {
  position: absolute;
  top: -6px;
  bottom: 0;
  width: 3px;
  margin-left: -1px;
}

.mark-deployed {
  left: 4%;
  background: oklch(0.25 0.03 55);
}

.mark-announced {
  left: 68%;
  background: oklch(0.55 0.19 27);
}

.bar-axis {
  position: relative;
  height: 38px;
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.35;
  color: oklch(0.47 0.04 55);
}

.bar-axis span {
  position: absolute;
  top: 0;
  white-space: nowrap;
}

.at-deployed {
  left: 4%;
}

.at-announced {
  left: 68%;
  transform: translateX(-50%);
  font-weight: 600;
  color: oklch(0.47 0.19 27);
}

.at-today {
  right: 0;
}

.vul-tab {
  position: absolute;
  left: 38%;
  right: -60px;
  top: 0;
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
  padding: 20px 96px 16px 48px;
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

.summary {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 22px;
  padding: 14px 96px 14px 48px;
  border-bottom: 1px solid oklch(0.91 0.015 80);
  font-size: 15px;
  line-height: 1.55;
  color: oklch(0.35 0.03 55);
}

.summary span {
  display: flex;
  align-items: center;
  gap: 8px;
}

.summary i {
  width: 9px;
  height: 9px;
  border-radius: 50%;
}

.row {
  display: grid;
  grid-template-columns: 132px 128px minmax(0, 1fr) auto;
  gap: 0 16px;
  align-items: center;
  padding: 13px 96px 13px 48px;
  border-top: 1px solid oklch(0.91 0.015 80);
  font-size: 15px;
  line-height: 1.55;
  color: oklch(0.45 0.04 55);
  white-space: nowrap;
}

.summary + .row {
  border-top: 0;
}

.row b {
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 14.5px;
  font-weight: 500;
  color: oklch(0.25 0.03 55);
}

.row-source {
  display: flex;
  align-items: center;
  gap: 7px;
  color: oklch(0.35 0.03 55);
}

.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 13px;
  font-style: normal;
  font-weight: 600;
  white-space: nowrap;
}

.badge-critical {
  background: oklch(0.55 0.19 27 / .12);
  color: oklch(0.47 0.19 27);
}

.badge-high {
  background: oklch(0.65 0.17 45 / .15);
  color: oklch(0.45 0.13 45);
}

.badge-medium {
  background: oklch(0.78 0.14 85 / .25);
  color: oklch(0.42 0.09 80);
}

@media (max-width: 1239px) {
  .vul {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr);
    gap: 28px 40px;
    align-items: center;
    height: auto;
  }

  .vul-note,
  .vul-month,
  .vul-tab {
    position: static;
    width: auto;
  }

  .vul-tab {
    grid-column: 1 / -1;
    margin: 0 0 -40px;
    padding-bottom: 40px;
    border-radius: 18px 18px 0 0;
  }

  .tab-top,
  .summary,
  .row {
    padding-right: 24px;
    padding-left: 24px;
  }
}

@media (max-width: 767px) {
  .vul {
    grid-template-columns: minmax(0, 1fr);
    gap: 24px;
  }

  .tab-top,
  .summary,
  .row {
    padding-right: 20px;
    padding-left: 20px;
  }
}

@media (max-width: 640px) {
  .row {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 4px 14px;
    align-items: start;
    white-space: normal;
  }

  .row b {
    grid-column: 1 / span 2;
  }

  .row .badge {
    grid-row: 1;
    grid-column: 3;
  }

  .row-source {
    grid-row: 2;
    grid-column: 1;
    white-space: nowrap;
  }

  .row-version {
    grid-row: 2;
    grid-column: 2 / span 2;
  }
}

@media (max-width: 520px) {
  .at-today {
    display: none;
  }
}
</style>
