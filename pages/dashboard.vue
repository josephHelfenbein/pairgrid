<template>
    <div class="min-h-screen bg-background">

      <SignedOut>
        <RedirectToSignUp />
      </SignedOut>
      <div v-if="loading==false">
        <Tabs default-value="chat" class="w-full p-4">
          <TabsList class="grid w-full grid-cols-3">
            <TabsTrigger value="chat">Chat</TabsTrigger>
            <TabsTrigger value="networking">Networking</TabsTrigger>
            <TabsTrigger value="preferences">Preferences</TabsTrigger>
          </TabsList>
          
          <TabsContent value="chat">
            <ChatTab 
            @toast-update="toastUpdate"
            :user="user"
            />
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
        <div
        v-if="showCallPopup"
        class="fixed top-0 left-0 z-50 flex items-center justify-center w-full h-full bg-black bg-opacity-50">
            <div class="relative p-6 bg-white rounded-lg shadow-lg w-80">
                <h3 class="text-lg font-semibold">Incoming Call</h3>
                <p class="mt-2 text-sm">{{ callerName }} is calling...</p>
                <div class="flex justify-between items-center mt-4 space-x-4">
                    <button @click="acceptCall" class="px-4 py-2 text-white bg-green-500 rounded hover:bg-green-600">
                        Accept
                    </button>
                    <button @click="declineCall" class="px-4 py-2 text-white bg-red-500 rounded hover:bg-red-600">
                        Decline
                    </button>
                </div>
                <button @click="closePopup" class="absolute top-2 right-2 text-gray-400 hover:text-gray-600">
                    &times;
                </button>
            </div>
        </div>
      </div>
      <div v-else class="flex justify-center items-center h-screen">
        <Loader size="150px" />  
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
  import Loader from '@/components/Loader'
  import { useRuntimeConfig } from '#app'
  import Pusher from 'pusher-js'
  import { useSession } from '@clerk/vue'
  const loading = ref(true)

  const { user } = useUser();
  const token = ref(null);
  const { session } = useSession();
  const reactiveSession = ref(session);
  const callPusher = ref(null);
  const showCallPopup = ref(false);

  const callerName = ref('Unknown Caller');
  const acceptCall = () => {
      console.log('Call accepted');
      showCallPopup.value = false;
  };
  const declineCall = () => {
      console.log('Call declined');
      showCallPopup.value = false;
  };
  const triggerIncomingCall = (name) => {
      callerName.value = name;
      showCallPopup.value = true;
  }

  watch(reactiveSession, async (newSession, oldSession) => {
    if (newSession) {
      try {
        token.value = await newSession.getToken();
      } catch (error) {
        console.error("Error getting token:", error);
      }
    }
  }, { immediate: true });
  const pusherConfig = {
    appKey: useRuntimeConfig().public.pusherAppKey,
    cluster: "us2",
  }

  const subscribeToCalls = () => {
    if (!token.value || !user?.value?.id) {
      console.error("Cannot subscribe to calls: Missing token or user ID.");
      return;
    }
    callPusher.value = new Pusher(pusherConfig.appKey, {
      cluster: pusherConfig.cluster,
      authEndpoint: 'https://www.pairgrid.com/api/pusherauth/pusherauth',
      auth: {
        headers: {
          'Accept':'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
      },
    })
    const callChannel = callPusher.value.subscribe(`private-call-${user.value.id}`)
    callChannel.bind('incoming-call', (data) => {
      triggerIncomingCall(data.callerName || 'Unknown Caller');
    })
    callPusher.value.connection.bind('error', (err) => {
      console.error('Pusher connection error:', err);
      toastUpdate('Session not found, try again.');
    });
  }

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
    } catch(error){
      console.error('Error loading preferences:', error);
    }
  }

  watch(() => token.value, (newToken) => {
    if (newToken && callPusher.value == null) {
      subscribeToCalls();
    }
  });
  watch(() => user.value, (newUser) => {
    if (newUser) {
      loadPreferences();
      loading.value = false;
      if(callPusher.value == null) subscribeToCalls();
    }
  });

  onMounted(() => {
    if (user.value) {
      loadPreferences();
      loading.value = false;
    }
  });
</script>