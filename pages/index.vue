<template>
  <div class="min-h-screen bg-background">
    <section class="container mx-auto px-4 py-24 text-center">
      <Motion
        :initial="{ y: 20, opacity: 0 }"
        :animate="{ y: 0, opacity: 1 }"
        :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 0.3 }"
      >
        <h1
          class="text-3xl md:text-6xl tracking-tighter mb-12 flex flex-row typing-animation"
        >
          <span>Connect. Code.</span>
          <span class="text-violet-500"> Create.</span>
      </h1>
      </Motion>
      <Motion
        :initial="{ y: 20, opacity: 0 }"
        :animate="{ y: 0, opacity: 1 }"
        :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 0.5 }"
      >
        <p class="text-sm sm:text-base lg:text-lg text-foreground max-w-xl mx-auto mb-12">
          Find your perfect coding partner with our real-time matchmaking platform. Connect with developers who share your interests and build amazing things together. (This website is still under construction)
        </p>
      </Motion>
      <div class="flex justify-center gap-4">
        <NuxtLink to="/dashboard">
          <Motion
            :initial="{ y: 20, opacity: 0 }"
            :animate="{ y: 0, opacity: 1 }"
            :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 0.7 }"
          >
            <Button
              size="lg"
              class="pl-2 pr-2 md:pl-5 md:pr-5 outline outline-2 outline-violet-600 bg-violet-900"
            >
              <UserPlus class="md:mr-2 h-5 w-5" />
              Find Partners Now
            </Button>
          </Motion>
        </NuxtLink>
        <Motion
          :initial="{ y: 20, opacity: 0 }"
          :animate="{ y: 0, opacity: 1 }"
          :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 0.9 }"
        >
          <Button size="lg" variant="outline" class="pl-4 pr-4 md:pl-5 md:pr-5 outline outline-2 outline-slate-700">
            <Play class="md:mr-2 h-5 w-5" />
            Watch Demo
          </Button>
        </Motion>
      </div>
    </section>

    <section class="container mx-auto px-4 py-24">
      <Motion
        :initial="{ y: 20, opacity: 0 }"
        :animate="{ y: 0, opacity: 1 }"
        :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 1.1 }"
      >
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-12">
          Everything You Need to Collaborate
        </h2>
      </Motion>
      <div class="grid md:grid-cols-3 gap-8">
        <Motion
          :initial="{ y: 20, opacity: 0 }"
          :animate="{ y: 0, opacity: 1 }"
          :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 1.3 }"
        >
          <Card class="p-6">
            <CardHeader>
              <Users class="h-12 w-12 mb-4" />
              <CardTitle class="text-xl md:text-2xl md:leading-none">Smart Matchmaking</CardTitle>
            </CardHeader>
            <CardContent class="text-sm md:text-base">
              Connect with developers who share your interests, tech stack, and coding goals. Our matching system finds your perfect coding partners.
            </CardContent>
          </Card>
        </Motion>

        <Motion
          :initial="{ y: 20, opacity: 0 }"
          :animate="{ y: 0, opacity: 1 }"
          :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 1.5 }"
        >
          <Card class="p-6">
            <CardHeader>
              <MessageSquare class="h-12 w-12 mb-4" />
              <CardTitle class="text-xl md:text-2xl md:leading-none">Real-time Chat</CardTitle>
            </CardHeader>
            <CardContent class="text-sm md:text-base">
              Communicate seamlessly with built-in real-time chat. Share ideas, code snippets, and collaborate effectively with your coding partners.
            </CardContent>
          </Card>
        </Motion>

        <Motion
          :initial="{ y: 20, opacity: 0 }"
          :animate="{ y: 0, opacity: 1 }"
          :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 1.7 }"
        >
          <Card class="p-6">
            <CardHeader>
              <ScreenShare class="h-12 w-12 mb-4" />
              <CardTitle class="text-xl md:text-2xl md:leading-none">Voice & Screen Sharing</CardTitle>
            </CardHeader>
            <CardContent class="text-sm md:text-base">
              Jump into voice calls and share your screen for interactive coding sessions. Perfect for pair programming and code reviews.
            </CardContent>
          </Card>
        </Motion>
      </div>
    </section>

    <section class="container mx-auto px-4 py-24">
      <Motion
        :initial="{ y: 20, opacity: 0 }"
        :animate="{ y: 0, opacity: 1 }"
        :transition="{ type: 'spring', stiffness: 260, damping: 20, delay: 1.9 }"
      >
        <Card class="p-12 text-center rounded-lg shadow-xl">
          <CardHeader>
            <Rocket class="h-16 w-16 mx-auto mb-6" />
            <CardTitle class="text-3xl">Start Collaborating Today</CardTitle>
          </CardHeader>
          <CardContent class="text-sm sm:text-base md:text-lg">
            Build meaningful connections and work on impactful projects with like-minded developers. Get started for free and see the difference.
          </CardContent>
          <div class="mt-8 flex justify-center">
            <NuxtLink to="/dashboard">
              <Button
                size="lg"
                class="pl-8 pr-8 md:pl-5 md:pr-5 outline outline-2 outline-violet-600 bg-violet-900"
              >
                Get Started
              </Button>
            </NuxtLink>
          </div>
        </Card>
      </Motion>
    </section>


    <footer class="border-t py-12 mx-auto px-4 mt-12 pt-8 text-center text-sm text-muted-foreground">
      © 2025 PairGrid. All rights reserved.
    </footer>
  </div>
</template>



<script setup>
  import { Motion } from '@oku-ui/motion';
  import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
  import { Button } from '@/components/ui/button';
  import { Play, Rocket, Users, UserPlus, MessageSquare, ScreenShare } from 'lucide-vue-next';
  useSeoMeta({
    title: 'Home',
    ogTitle: 'PairGrid - Home',
    twitterTitle: 'PairGrid - Home',
  });
</script>

<style scoped>
  @keyframes typing {
    0% {
      max-width: 0;
    }
    100% {
      max-width: 100%;
    }
  }

  @keyframes blink {
    0%,
    100% {
      border-color: transparent;
    }
    50% {
      border-color: #fff;
    }
  }

  .typing-animation {
    display: inline-block;
    overflow: hidden;
    white-space: nowrap;
    max-width: 0;
    border-right: 3px solid #fff;
    animation: typing 3s steps(40) 1s forwards, blink 0.75s step-end 3s infinite;
  }
</style>