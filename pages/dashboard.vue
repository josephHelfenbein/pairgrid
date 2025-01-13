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
            @call-user="triggerOutgoingCall"
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
            <div v-if="callType=='incoming'" class="relative p-6 rounded-lg bg-black shadow-lg w-80">
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
            </div>
            <div v-if="callType=='outgoing'" class="relative p-6 rounded-lg bg-black shadow-lg w-80">
                <p class="mt-2 text-sm">Calling {{ callerName }}...</p>
                <div class="flex justify-between items-center mt-4 space-x-4">
                    <button v-if="callStatus=='calling'" @click="cancelCall" class="px-4 py-2 text-white bg-red-500 rounded hover:bg-red-600">
                        Cancel
                    </button>
                    <p v-else-if="callStatus=='declined'" class="text-sm">Call was declined.</p>
                    <p v-else-if="callStatus=='canceled'" class="text-sm">Call ended.</p>
                </div>
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
  const callType = ref(null);

  const callerName = ref('Unknown Caller');
  const callerID = ref(null);
  const callStatus = ref(null);
  const peerConnection = ref(null);
  const remoteAudio = ref(null);
  const acceptCall = async () => {
    try {
      console.log('Call accepted');
      callType.value = "outgoing";
      callStatus.value = "calling";

      peerConnection.value.onicecandidate = (event) => {
        if (event.candidate) {
          sendSignalingMessage('ice-candidate', { candidate: event.candidate });
        }
      };

      peerConnection.value.ontrack = (event) => {
        remoteAudio.value.srcObject = event.streams[0];
        remoteAudio.value.play();
      };

      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      stream.getTracks().forEach((track) => peerConnection.value.addTrack(track, stream));
      
      const offer = await peerConnection.value.createOffer();
      await peerConnection.value.setLocalDescription(offer);
      sendSignalingMessage('sdp-offer', { sdp: offer });
    } catch (err) {
      console.error('Error accepting call:', err);
      toastUpdate('Error accepting call');
    }
  };

  const sendSignalingMessage = async (type, data) => {
    try {
      if (!token.value) {
        console.error("Token not available");
        return;
      }
      const payload = {
        type: type,
        user_id: user.value.id,
        recipient_id: callerID.value,
        ...data,
      };
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error('Failed to send signaling message');
    } catch (err) {
      console.error('Error sending signaling message:', err);
      toastUpdate('Error sending signaling message, please try again.');
    }
  };

  const declineCall = async () => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const payload = {
        caller_id: callerID.value,
        callee_id: user.value.id,
        type: "decline",
      }
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      })
      if (!response.ok) throw new Error('Failed to decline call');
    } catch (err) {
      console.error(err)
      toastUpdate('Error declining call')
    }
    showCallPopup.value = false;
  };
  const cancelCall = async () => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const payload = {
        caller_id: user.value.id,
        callee_id: callerID.value,
        type: "cancel",
      }
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      })
      if (!response.ok) throw new Error('Failed to cancel call');
      peerConnection.value.getSenders().forEach(sender => {
        if (sender.track) sender.track.stop();
      });
      if (remoteAudio.value && remoteAudio.value.srcObject) {
        remoteAudio.value.srcObject.getTracks().forEach(track => track.stop());
      } else {
        console.error('No media stream found for remoteAudio.');
      }
      callStatus.value = "canceled";
      setTimeout(()=>{showCallPopup.value = false;}, 2500);
    } catch (err) {
      console.error(err)
      toastUpdate('Error cancelling call')
    }
  };
  const triggerIncomingCall = (name) => {
      callerName.value = name;
      showCallPopup.value = true;
      callType.value = "incoming";
  }
  const triggerOutgoingCall = (name, id) => {
      callerName.value = name;
      callerID.value = id;
      showCallPopup.value = true;
      callType.value = "outgoing";
      callStatus.value = "calling";
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
      if(showCallPopup.value == true){
        console.log('Call already in progress');
        return;
      }
      callerID.value = data.caller_id;
      triggerIncomingCall(data.caller_name || 'Unknown Caller');
    })
    callChannel.bind('decline-call', (data) => {
      if(data.caller_id == user.value.id){
        console.log('Call declined by user');
        callStatus.value = "declined";
        setTimeout(()=>{showCallPopup.value = false;}, 2500);
      }
    })
    callChannel.bind('cancel-call', (data) => {
      if((data.caller_id == callerID.value && callType.value == "incoming") || ((data.caller_id == callerID.value || data.callee_id == callerID.value) && callType.value == "outgoing")){
        console.log('Call canceled by user');
        peerConnection.value.getSenders().forEach(sender => {
          if (sender.track) sender.track.stop();
        });
        if (remoteAudio.value && remoteAudio.value.srcObject) {
          remoteAudio.value.srcObject.getTracks().forEach(track => track.stop());
        } else {
          console.error('No media stream found for remoteAudio.');
        }
        callStatus.value = "canceled";
        setTimeout(()=>{showCallPopup.value = false;}, 2500);
      }
    })
    callChannel.bind('webrtc-message', async (data) => {
      try {
        if (data.type === 'sdp-offer') {
          await peerConnection.value.setRemoteDescription(new RTCSessionDescription(data.sdp));

          const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
          stream.getTracks().forEach((track) => peerConnection.value.addTrack(track, stream));

          const answer = await peerConnection.value.createAnswer();

          await peerConnection.value.setLocalDescription(answer);

          sendSignalingMessage('sdp-answer', { sdp: answer });
        } else if (data.type === 'ice-candidate') {
          await peerConnection.value.addIceCandidate(new RTCIceCandidate(data.candidate));
        } else if (data.type === 'sdp-answer') {
          await peerConnection.value.setRemoteDescription(new RTCSessionDescription(data.sdp));

          peerConnection.value.ontrack = (event) => {
            remoteAudio.value.srcObject = event.streams[0];
            remoteAudio.value.play();
          };
        }
      } catch (error) {
        console.error('Error handling WebRTC signaling:', error);
      }
    });
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
    peerConnection.value = new RTCPeerConnection({
      iceServers: [
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun1.l.google.com:19302' },
      { urls: 'stun:stun2.l.google.com:19302' },
      ],
    });
    remoteAudio.value = new Audio();
  });
</script>