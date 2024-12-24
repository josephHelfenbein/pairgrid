<template>
    <div class="min-h-screen bg-background">
      
      <SignedOut>
        <RedirectToSignUp />
      </SignedOut>

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
          <PreferencesTab v-if="preferences !== null" :preferences="preferences" @update-preferences="updatePreferences" />
        </TabsContent>
      </Tabs>
    </div>
</template>
  
<script setup>
  import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
  import ChatTab from '@/components/ChatTab.vue'
  import NetworkingTab from '@/components/NetworkingTab.vue'
  import PreferencesTab from '@/components/PreferencesTab.vue'
  import { reactive, onMounted, watch } from 'vue'
  import { useUser } from '@clerk/vue'

  const { user } = useUser();
  const preferences = reactive({
    bio: '',
    language: [],
    specialty: '',
    interests: [],
    occupation: '',
  })
  const updatePreferences = (updatedPreferences) => {
    Object.assign(preferences, updatedPreferences)
  }
  async function loadPreferences(){
    try{
      if(!user.value){
        throw new Error('User not found');
      }
      const response = await fetch('https://www.pairgrid.com/api/getuser/getuser', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({id: user.value.id}),
      });
      if(!response.ok){
        throw new Error(`Failed to load preferences: ${response.statusText}`);
      }
      const data = await response.json();
      Object.assign(preferences, {
        bio: data.bio || '',
        language: data.language || [],
        specialty: data.specialty || '',
        interests: data.interests || [],
        occupation: data.occupation || '',
      })
      console.log('Preferences loaded successfully:', preferences.value);
    } catch(error){
      console.error('Error loading preferences:', error);
    }
  }

  watch(() => user.value, (newUser) => {
    if (newUser) {
      loadPreferences();
    }
  });

  onMounted(() => {
    if (user.value) {
      loadPreferences();
    }
  });
</script>