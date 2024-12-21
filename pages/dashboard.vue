<template>
    <div class="min-h-screen bg-background">
      <header class="bg-primary text-primary-foreground p-4">
        <h1 class="text-2xl font-bold">My Dashboard</h1>
      </header>
      
      <Tabs default-value="chat" class="w-full p-4">
        <TabsList class="grid w-full grid-cols-3">
          <TabsTrigger value="chat">Chat</TabsTrigger>
          <TabsTrigger value="networking">Networking</TabsTrigger>
          <TabsTrigger value="preferences">Preferences</TabsTrigger>
        </TabsList>
        
        <TabsContent value="chat">
          <ChatTab />
        </TabsContent>
        <TabsContent value="networking">
          <NetworkingTab />
        </TabsContent>
        <TabsContent value="preferences">
          <PreferencesTab />
        </TabsContent>
      </Tabs>
    </div>
  </template>
  
  <script setup>
  import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
  import ChatTab from '@/components/ChatTab.vue'
  import NetworkingTab from '@/components/NetworkingTab.vue'
  import PreferencesTab from '@/components/PreferencesTab.vue'
  </script>