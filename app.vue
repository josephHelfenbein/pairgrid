<script setup lang="ts"></script>
<template>
  <header class="sticky z-40 top-0 backdrop-blur-lg border-b border-border flex justify-between items-center h-14 pr-4 pl-4">
    <NuxtLink to="/">PairGrid</NuxtLink>
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