<script setup lang="ts"></script>
<template>
  <header class="sticky z-40 top-0 backdrop-blur-lg border-b border-border flex justify-between items-center h-14 pr-4 pl-4">
      <NuxtLink to="/">PairGrid</NuxtLink>
      <div>
          <SignedOut>
              <SignInButton />
          </SignedOut>
          <SignedIn>
              <UserButton />
          </SignedIn>
      </div>
  </header>
  <NuxtPage />
</template>