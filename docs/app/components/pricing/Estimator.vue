<script setup lang="ts">
const appUrl = '/app';

const { groups, estimate } = usePricingEstimate();
const first = groups[0]!;
const rest = groups.slice(1);
</script>

<template>
  <section id="estimate" class="estimator mx-auto max-w-content scroll-mt-28 px-6" aria-labelledby="est">
    <LandingChapterHeading id="est" beat="grow" promise="Your bill" rest="worked out in advance.">
      Describe what you would run, in the sizes you pick in the dashboard. The estimate uses the rates above and the same rule as the invoice.
    </LandingChapterHeading>

    <LandingStage tint="grow" class="estimator-stage mt-10 px-4 pt-7 pb-8 sm:mt-14 sm:px-9 sm:pt-10 sm:pb-11">
      <div class="grid items-start gap-x-16 gap-y-8 lg:grid-cols-[1.25fr_1fr]">
        <div class="grid min-w-0 gap-7">
          <div>
            <p class="mb-2.5 text-base font-semibold text-[oklch(0.25_0.03_55)] dark:text-stone-100">{{ first.label }}</p>
            <LandingTabs v-model="first.selected.value" :options="first.options" :label="first.label" />
          </div>
          <div class="grid min-w-0 gap-7">
            <div v-for="group in rest" :key="group.label">
              <p class="mb-2.5 text-base font-semibold text-[oklch(0.25_0.03_55)] dark:text-stone-100">{{ group.label }}</p>
              <LandingTabs v-model="group.selected.value" :options="group.options" :label="group.label" />
            </div>
            <p class="max-w-[640px] text-[15px] text-[oklch(0.38_0.05_60)] dark:text-stone-300">
              Assumes each instance gets half a vCPU per GB of memory, uses {{ estimate.memoryShare }} of its memory and {{ estimate.cpuShare }} of its CPU on average, and that a database runs as two instances that hold half of their memory. Your app will differ. Billing in the dashboard shows the real numbers as they add up.
            </p>
            <div class="reach" aria-hidden="true">
              <span>You pay<small>{{ estimate.plan }} plan</small></span>
              <b>{{ estimate.total }}</b>
            </div>
          </div>
        </div>

        <div class="out min-w-0" aria-live="polite">
          <div class="flex items-baseline justify-between gap-4 pb-3.5">
            <b class="text-lg font-semibold text-[oklch(0.25_0.03_55)]">A month of this</b>
            <span class="text-[14.5px] text-[oklch(0.47_0.04_55)]">{{ estimate.plan }} plan</span>
          </div>
          <div v-for="line in estimate.lines" :key="line.label" class="line">
            <span>{{ line.label }}<small>{{ line.detail }}</small></span>
            <span>{{ line.amount }}</span>
          </div>
          <div class="line font-semibold"><span>Usage</span><span>{{ estimate.usage }}</span></div>
          <div class="line"><span>{{ estimate.plan }} plan</span><span>{{ estimate.fee }}</span></div>
          <div class="line included"><span>Usage included</span><span>{{ estimate.credit }}</span></div>
          <div class="total"><b>You pay</b><span>{{ estimate.total }}</span></div>
          <p class="mt-3.5 text-[15px]/[1.45] text-[oklch(0.45_0.04_55)]">{{ estimate.note }}</p>
          <p class="mt-5.5">
            <a
              :href="`${appUrl}/login`"
              class="flex min-h-15 w-full items-center justify-center rounded-lg border border-transparent bg-primary px-7 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600"
            >
              Start with {{ estimate.plan }}
            </a>
          </p>
        </div>
      </div>
    </LandingStage>
  </section>
</template>

<style scoped>
.estimator .estimator-stage {
  overflow: clip;
}

.estimator-stage :deep(.tab) {
  --tab-off: oklch(0.36 0.05 60);
}

.dark .estimator-stage :deep(.tab) {
  --tab-off: oklch(0.84 0.04 66);
}

.out,
.reach {
  box-sizing: border-box;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
  color: oklch(0.25 0.03 55);
}

.out {
  padding: 28px 32px 26px;
}

.reach {
  position: sticky;
  bottom: 12px;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 18px;
  font-size: 16px;
  font-weight: 600;
}

.reach small {
  display: block;
  font-size: 13.5px;
  font-weight: 400;
  color: oklch(0.47 0.04 55);
}

.reach b {
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  line-height: 1;
  white-space: nowrap;
}

.line {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-top: 1px solid oklch(0.91 0.015 80);
  font-size: 16px;
}

.line small {
  display: block;
  font-size: 14px;
  font-weight: 400;
  color: oklch(0.47 0.04 55);
}

.line > span:last-child {
  font-family: var(--font-mono);
  font-size: 15px;
  white-space: nowrap;
}

.included {
  color: oklch(0.43 0.12 160);
}

.total {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 4px 16px;
  margin-top: 4px;
  padding-top: 18px;
  border-top: 2px solid oklch(0.25 0.03 55);
}

.total b {
  font-size: 17px;
  font-weight: 600;
}

.total span {
  font-family: var(--font-display);
  font-size: 56px;
  line-height: 1;
  white-space: nowrap;
}

@media (min-width: 1024px) {
  .reach {
    display: none;
  }
}

@media (max-width: 700px) {
  .out {
    padding: 22px 20px;
  }

  .total span {
    font-size: 44px;
  }
}
</style>
