<script setup lang="ts">
</script>

<template>
  <section class="px-6">
    <div class="relative mx-auto max-w-450 -rotate-[0.5deg] px-10 py-9 shadow-xl shadow-neutral-900/5 md:px-10 lg:px-22 2xl:px-42 md:py-20 bg-white md:bg-[linear-gradient(90deg,#ffffff_0%,#ffffff_50%,#F7F7F7_50%,#ffffff_50.5%,#ffffff_100%)]">
      <div class="pointer-events-none absolute inset-y-0 left-1/2 hidden w-32 -translate-x-1/2 md:block bg-[linear-gradient(to_right,transparent,rgba(49,30,4,0.06)_49%,rgba(49,30,4,0.09)_50%,rgba(49,30,4,0.06)_55%,transparent)]" />
      <div class="relative grid gap-12 md:grid-cols-2 md:gap-20 lg:gap-44 2xl:gap-84">
        <div class="space-y-6 text-justify font-light leading-[1.7] text-neutral-900 text-lg lg:text-xl 2xl:text-2xl">
          <p>In my last 13 years as a software engineer I’ve experienced many ways to deploy software.</p>
          <p>In the good ol’ days I copied files via FTP to our server running under the desk. This was tedious and painful. But I’ve always appreciated the control I had over the infrastructure. <strong class="font-bold text-pink-500">It was truly mine.</strong></p>
          <p>Years after that I stumbled across Heroku and Vercel. Simply connecting your repository and watching how your code was shipped felt nothing short of magical. <strong class="font-bold text-violet-400">It was a joy to use.</strong></p>
          <p>But soon I realized that this simplicity came at the cost of giving up control and sovereignty.</p>
          <p>In my next job I was introduced to containers and, eventually, Kubernetes. A fascinating system designed to handle thousands of containers. It was immediately obvious: <strong class="font-bold text-red-400">It was built to grow.</strong></p>
        </div>

        <div class="flex h-full flex-col">
          <div class="space-y-6 text-justify font-light leading-[1.7] text-neutral-900 text-lg lg:text-xl 2xl:text-2xl">
            <p>But that focus on scalability made it too complicated for most software developers who simply wanted to ship software.</p>
            <p class="font-bold">I’ve built Lucity to combine the best parts of each of those solutions into one platform.</p>
            <p>So sign up, deploy your first app and let me know what you think of it, after you’ve used it for a bit.</p>
          </div>

          <div class="mt-auto pt-12">
            <img src="/img/signature-christian.png" alt="Christian" class="h-16 w-auto">
            <div class="mt-4 flex items-center gap-3">
              <img src="/img/founder-avatar.png" alt="Christian Blättler" width="48" height="48" class="size-12 rounded-full">
              <div class="text-lg leading-snug lg:text-xl text-neutral-900">
                <p>
                  <span class="font-bold">Christian Blättler, </span>
                  <a href="mailto:christian@zeitlos.software" class="font-bold text-blue-600 hover:underline">christian@zeitlos.software</a>
                </p>
                <p class="font-light italic text-neutral-600">Creator of Lucity</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
