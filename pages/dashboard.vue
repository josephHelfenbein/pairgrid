<template>
    <div class="min-h-screen bg-background">

      <SignedOut>
        <RedirectToSignUp />
      </SignedOut>

      <BounceLoader v-if="loading" color="primary" size="60px" class="m-auto mt-20" />

      <div v-else>
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
            <NetworkingTab 
            @toast-update="toastUpdate"
            :user="user"
            />
          </TabsContent>
          <TabsContent value="preferences">
            <PreferencesTab v-if="preferences !== null" :preferences="preferences" 
            :user="user"
            @update-preferences="updatePreferences" />
          </TabsContent>
        </Tabs>
        <Toaster />
      </div>
    </div>
</template>
  
<script setup>
  import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
  import ChatTab from '@/components/ChatTab.vue'
  import NetworkingTab from '@/components/NetworkingTab.vue'
  import PreferencesTab from '@/components/PreferencesTab.vue'
  import { reactive, onMounted, watch } from 'vue'
  import { useUser } from '@clerk/vue'
  import { useToast } from '@/components/ui/toast/use-toast'
  import { BounceLoader } from '@saeris/vue-spinners';
  const loading = ref(true);

  const { user } = useUser();
  const { toast } = useToast();
  const preferences = reactive({
    bio: '',
    language: [],
    specialty: '',
    interests: [],
    occupation: '',
  })
  const updatePreferences = (updatedPreferences) => {
    Object.assign(preferences, updatedPreferences);
    toast({description: 'Saved preferences.'});
  }
  const toastUpdate = (message) => {
    toast({description: message});
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
      console.log('Preferences loaded successfully:', preferences);
    } catch(error){
      console.error('Error loading preferences:', error);
    }
  }

  watch(() => user.value, (newUser) => {
    if (newUser) {
      loading.value = false;
      loadPreferences();
    }
  });

  onMounted(() => {
    if (user.value) {
      loadPreferences();
    }
    else loading.value = false;
  });
</script>