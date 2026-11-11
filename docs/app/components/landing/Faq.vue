<script setup lang="ts">
const questions = [
  { q: 'What can I run on Lucity?', a: 'Web apps, APIs and background workers in whatever you write them in: Node.js, Python, Go, PHP, Ruby, Java, Rust, Elixir, .NET and more. Lucity builds them straight from your repository, and it runs any public container image too.' },
  { q: 'Is it ready for production?', a: 'Yes. Databases run on two machines with automatic failover and continuous backups, rollouts happen without downtime, and a failed deploy never replaces a version that works.' },
  { q: 'What does it cost?', a: 'Hobby is CHF 5 a month and Pro is CHF 25, and each includes that much usage. Beyond it you pay for the CPU, memory and storage your apps actually use. The first 14 days are free.' },
  { q: 'Where does my data live?', a: 'In the EU. Apps and databases run in Germany, files and backups in France, and the company behind Lucity is Swiss.' },
  { q: 'What happens if I want to leave?', a: 'Click Eject. You download your whole setup in a standard format and run it on your own servers or another cloud, without Lucity. Copying your data over is manual work, and the guide walks you through it.' },
  { q: 'Can I host Lucity myself?', a: 'Yes. The whole platform is open source and installs on your own infrastructure, for free.' },
];
</script>

<template>
  <FaqSection
    id="faq"
    title="Before you ask"
    :items="questions"
    ask-label="Ask us directly"
    ask-href="mailto:christian@zeitlos.software"
  />
</template>
