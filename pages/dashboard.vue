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
            <TabsTrigger value="preferences">Profile</TabsTrigger>
          </TabsList>
          
          <TabsContent value="chat">
            <ChatTab 
            @toast-update="toastUpdate"
            @call-user="triggerOutgoingCall"
            :user="user"
            :preferences="preferences"
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
          class="popup-window fixed z-50 bg-gray-800 text-white w-72 shadow-lg rounded-lg overflow-hidden"
          :style="{ top: popupTop + 'px', left: popupLeft + 'px', width: popupWidth + 'px', height: popupHeight + 'px' }"
          ref="callPopup"
        >
          <div class="bg-gray-900 p-4 flex items-center justify-between cursor-grab"
            @mousedown="startDrag"  
            @touchstart="startDrag"
          >
            <h3 class="text-lg font-semibold">
              {{ callType === 'incoming' ? ((screenshareEnabled) ? 'Incoming Screenshare Call': 'Incoming Voice Call') : 'Call Progress' }}
            </h3>
          </div>

          <div class="p-6">
            <div v-if="callType === 'incoming'">
              <p class="text-center text-sm">{{ callerName }} is calling...</p>
              <div class="flex justify-center space-x-4 mt-4">
                <button
                  @click="acceptCall"
                  class="px-4 py-2 bg-green-500 text-white rounded-full hover:bg-green-600"
                >
                  Accept
                </button>
                <button
                  @click="declineCall"
                  class="px-4 py-2 bg-red-500 text-white rounded-full hover:bg-red-600"
                >
                  Decline
                </button>
              </div>
            </div>

            <div v-else-if="callStatus === 'calling'">
              <p class="text-center text-sm">Calling {{ callerName }}...</p>
              <div class="flex justify-center mt-4">
                <button
                  @click="cancelCall"
                  class="px-4 py-2 bg-red-500 text-white rounded-full hover:bg-red-600"
                >
                  Cancel
                </button>
              </div>
            </div>

            <div v-else-if="callStatus === 'active'">
              <p class="text-center text-sm">Talking to {{ callerName }}</p>
              <p class="text-center mt-2 text-sm">Duration: {{ callDuration }}</p>
              <div class="flex justify-center space-x-4 mt-4">
                <button
                  @click="toggleScreenshare"
                  class="px-4 py-2 bg-blue-500 text-white rounded-full hover:bg-blue-600"
                >
                  {{ showLocal ? 'Disable Screenshare' : 'Enable Screenshare' }}
                </button>
                <button
                  @click="cancelCall"
                  class="px-4 py-2 bg-red-500 text-white rounded-full hover:bg-red-600"
                >
                  End Call
                </button>
              </div>
            </div>
            

            <div v-else-if="callStatus === 'declined'">
              <p class="text-center text-sm">Call was declined by {{ callerName }}.</p>
            </div>
            <div v-else-if="callStatus === 'taken'">
              <p class="text-center text-sm">{{ callerName }} is already on a call.</p>
            </div>
            <div v-else-if="callStatus === 'canceled'">
              <p class="text-center text-sm">Call was canceled.</p>
            </div>
            <div class="mt-6">
              <div v-if="showLocal||(showRemote&&callStatus==='active'&&callType!=='incoming')" class="w-full h-full bg-black rounded-lg">
                <video
                  ref="remoteScreen"
                  v-bind:class="(showRemote)?'w-full h-full object-contain':'hidden'"
                  autoplay
                  muted
                ></video>
                <video
                  ref="localScreen"
                  v-bind:class="(showLocal)?((!showRemote)?'w-full h-full object-contain':'absolute bottom-2 right-2 w-24 h-16 object-contain border-2 border-white rounded'):'hidden'"
                  autoplay
                  muted
                ></video>
              </div>
            </div>
            <div class="resize-handle absolute bottom-0 right-0 cursor-se-resize p-1"
              @mousedown="startResize"
              @touchstart="startResize"
            ></div>
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
  import { reactive, onMounted, watch, nextTick } from 'vue'
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
  const popupTop = ref(0);
  const popupLeft = ref(0);
  const popupWidth = ref(300);
  const popupHeight = ref(200);
  const isDragging = ref(false);
  const isResizing = ref(false);
  const callDuration = ref('00:00');
  const screenshareEnabled = ref(false);
  const localScreen = ref(null);
  const remoteScreen = ref(null);
  const showLocal = ref(false);
  const showRemote = ref(false);
  let callStartTime = null;
  let callInterval = null;

  useSeoMeta({
    title: 'Dashboard',
    description: 'Manage your coding partnerships with PairGrid. Collaborate, chat, and connect with developers in real time.',
    ogTitle: 'PairGrid - Dashboard',
    twitterTitle: 'PairGrid - Dashboard',
  });
  const toggleScreenshare = () => {
    if (showLocal.value) {
      disableScreenshare();
    } else {
      enableScreenshare();
    }
  };
  const enableScreenshare = async() =>{
    try{
      screenshareEnabled.value = true;
      popupHeight.value = 400;
      sendSignalingMessage('enableScreenshare', {});
      if(!peerConnection.value || peerConnection.value.connectionState === 'closed') {
        peerConnection.value = new RTCPeerConnection({
          iceServers: [
          { urls: 'stun:stun.l.google.com:19302' },
          { urls: 'stun:stun1.l.google.com:19302' },
          { urls: 'stun:stun2.l.google.com:19302' },
          ],
        });
        peerConnection.value.onicecandidate = (event) => {
          if (event.candidate) {
            sendSignalingMessage('ice-candidate', { candidate: event.candidate });
          }
        };
        peerConnection.value.ontrack = handleTracks;
      }
      let stream;
      console.log('Initializing screen sharing...');
      showLocal.value=true;
      const displayStream = await navigator.mediaDevices.getDisplayMedia({ video: true })
      const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true })
      
      stream = new MediaStream([
        ...displayStream.getVideoTracks(),
        ...audioStream.getAudioTracks()
      ])

      if (localScreen.value) {
        localScreen.value.srcObject = stream
        await localScreen.value.play()
      }

      stream.getTracks().forEach((track) => {
        console.log(`Adding track: ${track.kind}`);
        peerConnection.value.addTrack(track, stream);
      });
      
      const offer = await peerConnection.value.createOffer();
      await peerConnection.value.setLocalDescription(offer);
      sendSignalingMessage('sdp-offer', { sdp: offer });
    } catch(error) {
      console.error('Error enabling screenshare:', error);
      toastUpdate('Error enabling screenshare');
    }
  }
  const disableScreenshare = async()=>{
    showLocal.value=false;
    popupHeight.value = 200;
    const stream = localScreen.value.srcObject;
    if(stream) stream.getTracks().forEach(track => track.stop());
    localScreen.value.srcObject = null;
    screenshareEnabled.value = false;
  }

  const centerPopup = () => {
    const popup = document.querySelector('.popup-window');
    if (popup) {
      popupWidth.value = 300;
      if(!screenshareEnabled.value) popupHeight.value = 200;
      else popupHeight.value = 400;

      const viewportHeight = window.innerHeight;
      const viewportWidth = window.innerWidth;

      const popupOffsetH = popup.offsetHeight;
      const popupOffsetW = popup.offsetWidth;

      popupTop.value = (viewportHeight - popupOffsetH) / 2;
      popupLeft.value = (viewportWidth - popupOffsetW) / 2;
    }
  };
  const startCallTimer = () => {
    callStartTime = new Date();
    callInterval = setInterval(() => {
      const elapsedTime = Math.floor((new Date() - callStartTime) / 1000);
      const minutes = Math.floor(elapsedTime / 60)
        .toString()
        .padStart(2, '0');
      const seconds = (elapsedTime % 60).toString().padStart(2, '0');
      callDuration.value = `${minutes}:${seconds}`;
    }, 1000);
  };

  const stopCallTimer = () => {
    clearInterval(callInterval);
    callInterval = null;
    callDuration.value = '00:00';
  };

  const startResize = (event) => {
    event.preventDefault();
    isResizing.value = true;
    const popup = event.target.closest('.popup-window');
    const startWidth = popup.offsetWidth;
    const startHeight = popup.offsetHeight;
    const startX = event.touches ? event.touches[0].clientX : event.clientX;
    const startY = event.touches ? event.touches[0].clientY : event.clientY;
    const moveHandler = (e) => {
      if(!isResizing.value) return;
      const clientX = e.touches?.[0]?.clientX || e.clientX;
      const clientY = e.touches?.[0]?.clientY || e.clientY;
      const newWidth = Math.max(startWidth + clientX - startX, 200);
      const newHeight = Math.max(startHeight + clientY - startY, 200);
      popupWidth.value = newWidth;
      popupHeight.value = newHeight;
    };
    const stopHandler = () => {
      isResizing.value = false;
      window.removeEventListener('mousemove', moveHandler);
      window.removeEventListener('mouseup', stopHandler);
      window.removeEventListener('touchmove', moveHandler);
      window.removeEventListener('touchend', stopHandler);
    }
    window.addEventListener('mousemove', moveHandler);
    window.addEventListener('mouseup', stopHandler);
    window.addEventListener('touchmove', moveHandler);
    window.addEventListener('touchend', stopHandler);
  };

  const startDrag = (event) => {
    isDragging.value = true;
    const popup = event.target.closest('.popup-window'); 
    const disableScroll = (e) => e.preventDefault();
    const rect = popup.getBoundingClientRect();
    const { clientX, clientY } = event.touches ? event.touches[0] : event;
    const offsetX = clientX - rect.left;
    const offsetY = clientY - rect.top;
    const isMobile = window.innerWidth <= 768 || /Mobi|Android/i.test(navigator.userAgent);
    const moveHandler = (e) => {
      if (!isDragging.value) return;
      const clientX = e.touches?.[0]?.clientX || e.clientX;
      const clientY = e.touches?.[0]?.clientY || e.clientY;

      const newTop = clientY - offsetY;
      const newLeft = clientX - offsetX;

      const viewportHeight = window.innerHeight;
      const viewportWidth = window.innerWidth;

      popupTop.value = Math.min(Math.max(newTop, 0), viewportHeight - popup.offsetHeight);
      popupLeft.value = Math.min(Math.max(newLeft, 0), viewportWidth - popup.offsetWidth);
    };
    const stopDrag = () => {
      isDragging.value = false;
      window.removeEventListener('mousemove', moveHandler);
      window.removeEventListener('mouseup', stopDrag);
      window.removeEventListener('touchmove', moveHandler);
      window.removeEventListener('touchend', stopDrag);
      if (isMobile) {
        document.body.style.overflow = '';
        window.removeEventListener('touchmove', disableScroll, { passive: false });
      }
    };
    if (isMobile) {
      document.body.style.overflow = 'hidden';
      window.addEventListener('touchmove', disableScroll, { passive: false });
    }

    window.addEventListener('mousemove', moveHandler);
    window.addEventListener('mouseup', stopDrag);
    window.addEventListener('touchmove', moveHandler);
    window.addEventListener('touchend', stopDrag);
  };
  
  const cleanupWebRTC = () => {
    if (peerConnection.value) {
      peerConnection.value.getSenders().forEach(sender => {
        if (sender.track) {
          sender.track.stop()
        }
      })
      peerConnection.value.close()
    }
    if (remoteAudio.value) {
      remoteAudio.value.srcObject = null
    }
    if (remoteScreen.value) {
      remoteScreen.value.srcObject = null
    }
  }

  const acceptCall = async () => {
    try {
      console.log('Call accepted');
      callType.value = "outgoing";
      callStatus.value = "active";
      startCallTimer();

      if(!peerConnection.value || peerConnection.value.connectionState === 'closed') {
        peerConnection.value = new RTCPeerConnection({
          iceServers: [
          { urls: 'stun:stun.l.google.com:19302' },
          { urls: 'stun:stun1.l.google.com:19302' },
          { urls: 'stun:stun2.l.google.com:19302' },
          ],
        });
        peerConnection.value.onicecandidate = (event) => {
          if (event.candidate) {
            sendSignalingMessage('ice-candidate', { candidate: event.candidate });
          }
        };
        peerConnection.value.ontrack = handleTracks;
      }

      if(screenshareEnabled.value) sendSignalingMessage('enableScreenshare', {});
      else sendSignalingMessage('disableScreenshare', {});

      let stream;
      if (screenshareEnabled.value) {
        console.log('Initializing screen sharing...');
        showLocal.value=true;
        const displayStream = await navigator.mediaDevices.getDisplayMedia({ video: true })
        const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true })
        stream = new MediaStream([
          ...displayStream.getVideoTracks(),
          ...audioStream.getAudioTracks()
        ])

        if (localScreen.value) {
          localScreen.value.srcObject = stream
          await localScreen.value.play()
        }
      } else {
        console.log('Initializing microphone...');
        stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      }

      stream.getTracks().forEach((track) => {
        console.log(`Adding track: ${track.kind}`);
        peerConnection.value.addTrack(track, stream);
      });
      
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
      showCallPopup.value = false;
      callStatus.value = "declined";
      showLocal.value = false;
      showRemote.value = false;
      stopCallTimer();
      setTimeout(()=>{showCallPopup.value = false;}, 2500);
    } catch (err) {
      console.error(err)
      toastUpdate('Error declining call')
    }
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
      showLocal.value = false;
      showRemote.value = false;
      stopCallTimer();
      cleanupWebRTC();
      disableScreenshare();
      setTimeout(()=>{showCallPopup.value = false;}, 2500);
    } catch (err) {
      console.error(err)
      toastUpdate('Error cancelling call')
    }
  };
  const triggerIncomingCall = (name, type) => {
      callerName.value = name;
      showCallPopup.value = true;
      callType.value = "incoming";
      if(type == "screen") {
        popupHeight.value = 400;
        screenshareEnabled.value = true;
      }
      else screenshareEnabled.value = false;
  }
  const triggerOutgoingCall = async (name, id, type) => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      if(showCallPopup.value) {
        toastUpdate('You are already on a call');
        return;
      }
      callerName.value = name;
      callerID.value = id;
      showCallPopup.value = true;
      callType.value = "outgoing";
      callStatus.value = "calling";
      if(!peerConnection.value || peerConnection.value.connectionState === 'closed') {
        peerConnection.value = new RTCPeerConnection({
          iceServers: [
          { urls: 'stun:stun.l.google.com:19302' },
          { urls: 'stun:stun1.l.google.com:19302' },
          { urls: 'stun:stun2.l.google.com:19302' },
          ],
        });
        peerConnection.value.onicecandidate = (event) => {
          if (event.candidate) {
            sendSignalingMessage('ice-candidate', { candidate: event.candidate });
          }
        };
      }
      const payload = {
        caller_id: user.value.id,
        callee_id: id,
        type: type,
        caller_name: user.value.fullName,
      }
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      })
      if(type=="screen") {
        try{
          showLocal.value=true;
          const mediaStream = await navigator.mediaDevices.getDisplayMedia({ video: true });
          if (localScreen.value) localScreen.value.srcObject = mediaStream;
          screenshareEnabled.value = true;
          popupHeight.value = 400;
          sendSignalingMessage('enableScreenshare', {});
        } catch(error){
          console.error('Error starting screenshare:', error);
          toastUpdate('Error starting screenshare');
        }
      }
      if (!response.ok) throw new Error('Failed to call user')
    } catch (err) {
      console.error(err)
      toastUpdate('Error calling user')
    }
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

  const handleTracks = async (event) => {
    console.log('Received tracks:', event);
    const track = event.track;
    if (track.kind === "audio") {
      remoteAudio.value.srcObject = new MediaStream([track]);
      if (remoteAudio.value.setSinkId) {
        await remoteAudio.value.setSinkId('default');
      }
      await remoteAudio.value.play();
      console.log('Remote audio track set.');
    } else if (track.kind === "video") {
      showRemote.value = true;
      remoteScreen.value.srcObject = new MediaStream([track]);
      await remoteScreen.value.play();
      console.log('Remote video track received and set on remoteScreen:', track);
    } else {
      console.warn('Unexpected track kind:', track.kind);
    }
  };
  let pendingCandidates = [];

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
    callChannel.bind('taken-call', (data) => {
      if(data.caller_id == user.value.id){
        console.log('Call taken by user');
        callStatus.value = "taken";
        showLocal.value = false;
        showRemote.value = false;
        setTimeout(()=>{showCallPopup.value = false;}, 2500);
      }
    })
    callChannel.bind('incoming-call', (data) => {
      if(showCallPopup.value == true){
        const payload = {
          caller_id: data.caller_id,
          callee_id: user.value.id,
          type: "taken",
        }
        fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token.value}`,
          },
          body: JSON.stringify(payload),
        })
        toastUpdate(`Call missed from ${data.caller_name || 'Unknown Caller'}.`);
        return;
      }
      callerID.value = data.caller_id;
      triggerIncomingCall(data.caller_name || 'Unknown Caller', data.type);
    })
    callChannel.bind('decline-call', (data) => {
      if(data.caller_id == user.value.id){
        console.log('Call declined by user');
        callStatus.value = "declined";
        showLocal.value = false;
        showRemote.value = false;
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
        cleanupWebRTC();
        disableScreenshare();
        callStatus.value = "canceled";
        showLocal.value = false;
        showRemote.value = false;
        setTimeout(()=>{showCallPopup.value = false;}, 2500);
      }
    })
    callChannel.bind('webrtc-message', async (data) => {
      if(callStatus.value != "active"){
        callStatus.value = "active";
        startCallTimer();
      }
      try {
        if(data.type === 'enableScreenshare' && callType.value !== 'incoming') {
          showRemote.value = true;
          popupHeight.value = 400;
        }
        if(data.type === 'disableScreenshare') {
          showRemote.value = false;
          if (remoteScreen.value) {
            remoteScreen.value.srcObject = null
          }
          popupHeight.value = 200;
        }
        if (data.type === 'sdp-offer') {
          await peerConnection.value.setRemoteDescription(new RTCSessionDescription(data.sdp));
          for (const pending of pendingCandidates) {
            await peerConnection.value.addIceCandidate(pending);
          }
          pendingCandidates = [];
          peerConnection.value.ontrack = handleTracks;

          let stream;
          if (screenshareEnabled.value && localScreen.value.srcObject) {
            console.log('Requesting screen sharing with audio...');
            
            const displayStream = localScreen.value.srcObject;
            const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true });

            stream = new MediaStream([
              ...displayStream.getVideoTracks(),
              ...audioStream.getAudioTracks(),
            ]);

          } else {
            console.log('Requesting audio-only media...');
            stream = await navigator.mediaDevices.getUserMedia({ audio: true });
          }
            
          stream.getTracks().forEach((track) => peerConnection.value.addTrack(track, stream));
          const answer = await peerConnection.value.createAnswer();
          await peerConnection.value.setLocalDescription(answer);
          sendSignalingMessage('sdp-answer', { sdp: answer });
          
        } else if (data.type === 'ice-candidate') {
          const candidate = new RTCIceCandidate(data.candidate)
          if (peerConnection.value.remoteDescription) {
            await peerConnection.value.addIceCandidate(candidate)
          } else {
            pendingCandidates.push(candidate)
          }
        } else if (data.type === 'sdp-answer') {
          await peerConnection.value.setRemoteDescription(new RTCSessionDescription(data.sdp));

          peerConnection.value.ontrack = handleTracks;
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
    profilePicture: '',
  })
  const updatePreferences = (updatedPreferences) => {
    Object.assign(preferences, updatedPreferences);
    toast({description: 'Saved profile.'});
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
        throw new Error(`Failed to load profile: ${response.statusText}`);
      }
      const data = await response.json();
      Object.assign(preferences, {
        bio: data.bio || '',
        language: data.language || [],
        specialty: data.specialty || '',
        interests: data.interests || [],
        occupation: data.occupation || '',
        profilePicture: data.profile_picture || '',
      })
    } catch(error){
      console.error('Error loading profile:', error);
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
  watch(() => showCallPopup.value, (newValue) => {
    setTimeout(() => centerPopup(), 10); 
  });

  onMounted(async () => {
    if (user.value) {
      loadPreferences();
      loading.value = false;
    }
    await nextTick();
    peerConnection.value = new RTCPeerConnection({
      iceServers: [
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun1.l.google.com:19302' },
      { urls: 'stun:stun2.l.google.com:19302' },
      ],
    });
    peerConnection.value.onicecandidate = (event) => {
      if (event.candidate) {
        sendSignalingMessage('ice-candidate', { candidate: event.candidate });
      }
    };
    peerConnection.value.ontrack = handleTracks;
    remoteAudio.value = new Audio();
    window.addEventListener('resize', centerPopup);
  });
  onBeforeUnmount(()=>{
    if(callPusher.value) callPusher.value.disconnect();
    cleanupWebRTC();
    window.removeEventListener('resize', centerPopup);
  })
</script>