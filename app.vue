<script setup lang="ts"></script>
<template>
  <header class="sticky z-40 top-0 backdrop-blur-lg border-b border-border flex justify-between items-center h-14 pr-4 pl-4">
    <NuxtLink class="flex items-center" style="font-family: 'Nunito', Verdana, Geneva, Tahoma, sans-serif;" to="/">
      <img src="/pairgrid-icon.svg" class="h-8 w-auto pr-2 md:pl-2" />
      <p class="text-lg">PairGrid</p>
    </NuxtLink>

    <div class="flex">
    <ThemeToggle />
    <SignedOut>
        <SignInButton />
    </SignedOut>
    <SignedIn>
        <div class="flex space-x-4">
            <NuxtLink to="/dashboard">
                <Button variant="outline">Dashboard</Button>
            </NuxtLink>
            <UserButton />
        </div>
    </SignedIn>
    </div>
  </header>
  <NuxtPage />
</template>
<script setup>
  import { Analytics } from '@vercel/analytics/nuxt'
  import { SpeedInsights } from '@vercel/speed-insights/nuxt'
</script>