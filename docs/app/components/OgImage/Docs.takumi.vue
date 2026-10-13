<script lang="ts" setup>
const { title, description, headline } = defineProps<{ title?: string; description?: string; headline?: string }>();

const mark = '<svg xmlns="http://www.w3.org/2000/svg" width="49" height="35" viewBox="0 0 250 181"><path d="M 130 0 L 130 76 C 130 102.64, 114.4 112, 70 112 L 60 112 C 15.6 112, 0 121.36, 0 148 L 0 181 L 169 181 L 169 148 C 169 121.36, 178.23 112, 204.5 112 L 214.5 112 C 240.77 112, 250 102.64, 250 76 L 250 0 Z" fill="#8843db"/></svg>';
</script>

<template>
  <div class="w-full h-full flex flex-col justify-between bg-[#f6ede0] px-[80px] py-[64px]">
    <div class="flex items-center">
      <div class="flex w-[49px] h-[35px]" v-html="mark" />
      <span
        class="ml-[30px] text-[36px] leading-none text-[#301c0e]"
        style="font-family: 'Redaction 35'"
      >Lucity Docs</span>
    </div>

    <div class="flex flex-col">
      <p
        v-if="headline"
        class="m-0 mb-4 text-[22px] uppercase tracking-[0.08em] text-[#8843db]"
        style="font-family: 'Mona Sans'; font-variation-settings: 'wght' 600"
      >
        {{ headline }}
      </p>
      <h1
        v-if="title"
        class="m-0 mb-6 text-[72px] font-normal leading-[1.05] text-[#301c0e] max-w-[1000px]"
        style="font-family: 'Redaction 20'"
      >
        {{ title?.slice(0, 60) }}
      </h1>
      <p
        v-if="description"
        class="m-0 text-[28px] leading-[1.4] text-[#7a6250] max-w-[960px]"
        style="font-family: 'Mona Sans'; font-variation-settings: 'wght' 400"
      >
        {{ description?.slice(0, 200) }}
      </p>
    </div>
  </div>
</template>
