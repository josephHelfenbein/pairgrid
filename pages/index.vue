<template>
    <div class="min-h-screen bg-background">
      <section class="container mx-auto px-4 py-24 text-center">
        <h1 class="text-4xl md:text-6xl font-bold tracking-tighter mb-6">
          Connect. Code. Create.
        </h1>
        <p class="text-xl text-muted-foreground max-w-2xl mx-auto mb-8">
          Find your perfect coding partner with our real-time matchmaking platform. Connect with developers who share your interests and build amazing things together. (This website is still under construction)
        </p>
        <div class="flex justify-center gap-4">
            <NuxtLink to="/dashboard">
                <Button size="lg" class="bg-gradient-to-t from-primary to-violet-800 hover:from-primary hover:to-violet-500">
                    <UserPlus class="mr-2 h-5 w-5" />
                    Find Partners Now
                </Button>
            </NuxtLink>
            <Button size="lg" variant="outline">
                <Play class="mr-2 h-5 w-5" />
                Watch Demo
            </Button>
        </div>
      </section>

      <section class="container mx-auto px-4 py-24">
        <h2 class="text-3xl font-bold text-center mb-12">Everything You Need to Collaborate</h2>
        <div class="grid md:grid-cols-3 gap-8">
          <Card class="p-6">
            <CardHeader>
              <Users class="h-12 w-12 mb-4" />
              <CardTitle>Smart Matchmaking</CardTitle>
            </CardHeader>
            <CardContent>
              Connect with developers who share your interests, tech stack, and coding goals. Our matching system finds your perfect coding partners.
            </CardContent>
          </Card>
  
          <Card class="p-6">
            <CardHeader>
              <MessageSquare class="h-12 w-12 mb-4" />
              <CardTitle>Real-time Chat</CardTitle>
            </CardHeader>
            <CardContent>
              Communicate seamlessly with built-in real-time chat. Share ideas, code snippets, and collaborate effectively with your coding partners.
            </CardContent>
          </Card>
  
          <Card class="p-6">
            <CardHeader>
              <Video class="h-12 w-12 mb-4" />
              <CardTitle>Video & Screen Sharing</CardTitle>
            </CardHeader>
            <CardContent>
              Jump into video calls and share your screen for interactive coding sessions. Perfect for pair programming and code reviews.
            </CardContent>
          </Card>
        </div>
      </section>
  
      <section class="container mx-auto px-4 py-24">
        <Card class="p-12 text-center bg-gradient-to-t from-primary to-violet-900 text-primary-foreground">
          <CardHeader>
            <CardTitle class="text-3xl font-bold">Ready to Find Your Coding Partner?</CardTitle>
            <CardDescription class="text-xl text-primary-foreground/90">
              Use our smart matchmaking platform
            </CardDescription>
          </CardHeader>
          <CardContent class="mt-6">
            <NuxtLink to="/dashboard">
                <Button size="lg" variant="secondary">
                Get Started for Free
                </Button>
            </NuxtLink>
          </CardContent>
        </Card>
      </section>
  
      <footer class="border-t py-12 mx-auto px-4 mt-12 pt-8 text-center text-sm text-muted-foreground">
            © 2024 PairGrid. All rights reserved.
      </footer>
    </div>
  </template>
  
  <script setup>
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
  import { UserPlus, Play, Users, MessageSquare, Video } from 'lucide-vue-next'
  </script>