<template>
  <div class="flex gap-3">
    <div class="hidden md:block md:w-1/3">
      <Card class="h-full">
        <CardHeader>
          <CardTitle>Friends</CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea class="h-[calc(100vh-300px)]">
            <FriendsList 
              :friends="friends" 
              :requests="requests" 
              :notifications="notifications"
              :friendsLoading="friendsLoading"
              :selectedFriend="selectedFriend"
              @selectFriend="selectFriend"
              @selectRequest="selectRequest"
              @acceptRequest="acceptRequest"
              @denyRequest="denyRequest"
            />
          </ScrollArea>
        </CardContent>
      </Card>
    </div>

    <div class="md:hidden w-full">
      <div v-if="selectedFriend" class="h-full">
        <Card class="h-full">
          <CardHeader class="flex flex-row justify-between items-center">
            <button @click="deselectFriend" class="p-2">
              <ChevronLeft class="h-5 w-5" />
            </button>
            <CardTitle class="flex-shrink-0 text-lg flex items-center">
              {{ selectedFriend.name }}
            </CardTitle>
            <div class="flex flex-row">
              <button @click="callFriend" class="p-2">
                <Phone class="h-5 w-5" />
              </button>
              <button @click="shareScreen" class="p-2">
                <ScreenShare class="h-5 w-5" />
              </button>
              <FriendOptions 
                :selectedFriend="selectedFriend"
                @removeFriend="removeFriend"
              />
            </div>
          </CardHeader>
          <CardContent>
            <ChatArea 
              :selectedFriend="selectedFriend"
              :messages="messages"
              :chatLoading="chatLoading"
              :newMessage="newMessage"
              @sendMessage="sendMessage"
              @updateNewMessage="newMessage = $event"
            />
          </CardContent>
        </Card>
      </div>

      <div v-else-if="requestProfile" class="h-full">
          <Card class="h-full flex flex-col">
            <CardHeader class="flex flex-row justify-between items-center">
              <button @click="deselectFriend" class="p-2">
                <ChevronLeft class="h-6 w-6" />
              </button>
              <CardTitle class="flex-shrink-0 flex items-center">{{ requestProfile?.name }}</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea class="h-[calc(100vh-300px)]">
                <p><strong>Specialty:</strong> {{ requestProfile?.specialty }}</p>
                <p><strong>Occupation:</strong> {{ requestProfile?.occupation }}</p>
                <p><strong>Bio:</strong> {{ requestProfile?.bio }}</p>
                <div>
                  <strong>Languages:</strong>
                  <div class="flex flex-wrap space-x-2 text-sm">
                    <p v-for="language in requestProfile?.language" :key="language" class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2">
                      {{ language }}
                    </p>
                  </div>
                </div>
                <div>
                  <strong>Interests:</strong>
                  <div class="flex flex-wrap space-x-2 text-sm">
                    <p v-for="interest in requestProfile?.interests" :key="interest" class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2">
                      {{ interest }}
                    </p>
                  </div>
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </div> 

      <div v-else class="h-full">
        <Card class="h-full">
          <CardHeader>
            <CardTitle>Friends</CardTitle>
          </CardHeader>
          <CardContent>
            <ScrollArea class="h-[calc(100vh-300px)]">
              <FriendsList 
                :friends="friends" 
                :requests="requests" 
                :notifications="notifications"
                :friendsLoading="friendsLoading"
                :selectedFriend="selectedFriend"
                @selectFriend="selectFriend"
                @selectRequest="selectRequest"
                @acceptRequest="acceptRequest"
                @denyRequest="denyRequest"
              />
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>

    <Card class="hidden md:block md:w-2/3">
      <CardHeader class="flex flex-row justify-between items-center">
        <CardTitle v-if="selectedFriend" class="flex-shrink-0 flex items-center">
          {{selectedFriend.name}}
        </CardTitle>
        <CardTitle v-else-if="requestProfile" class="flex-shrink-0 flex items-center">
          {{requestProfile.name}}
        </CardTitle>
        <CardTitle v-else class="flex-shrink-0 flex items-center">
          Select a friend
        </CardTitle>
        <div v-if="selectedFriend" class="flex flex-row">
          <button @click="callFriend" class="p-2">
            <Phone class="h-5 w-5" />
          </button>
          <button @click="shareScreen" class="p-2">
            <ScreenShare class="h-5 w-5" />
          </button>
          <FriendOptions 
            :selectedFriend="selectedFriend"
            @removeFriend="removeFriend"
          />
        </div>
      </CardHeader>
      <CardContent>
        <ChatArea 
          v-if="selectedFriend"
          :selectedFriend="selectedFriend"
          :messages="messages"
          :chatLoading="chatLoading"
          :newMessage="newMessage"
          @sendMessage="sendMessage"
          @updateNewMessage="newMessage = $event"
        />
        <div v-else-if="requestProfile" class="flex justify-center items-center h-full">
          <div class="space-y-2">
            <p><strong>Specialty:</strong> {{ requestProfile?.specialty }}</p>
            <p><strong>Occupation:</strong> {{ requestProfile?.occupation }}</p>
            <p><strong>Bio:</strong> {{ requestProfile?.bio }}</p>
            <div>
              <strong>Languages:</strong>
              <div class="flex flex-wrap space-x-2 text-sm">
                <p v-for="language in requestProfile?.language" :key="language" class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2">
                  {{ language }}
                </p>
              </div>
            </div>
            <div>
              <strong>Interests:</strong>
              <div class="flex flex-wrap space-x-2 text-sm">
                <p v-for="interest in requestProfile?.interests" :key="interest" class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2">
                  {{ interest }}
                </p>
              </div>
            </div>
          </div>
        </div> 
        <div v-else class="flex justify-center items-center h-full">
          <p class="text-gray-500">Select a friend to start chatting</p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
  
<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import { ChevronLeft, Phone, ScreenShare } from 'lucide-vue-next'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { ScrollArea } from '@/components/ui/scroll-area'
  import { useRuntimeConfig } from '#app'
  import Pusher from 'pusher-js'
  import { useSession } from '@clerk/vue'

  import FriendsList from './FriendsList.vue'
  import FriendOptions from './FriendOptions.vue'
  import ChatArea from './ChatArea.vue'

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })

  const emit = defineEmits(['toast-update', 'call-user'])
  const token = ref(null);
  const { session } = useSession();
  const reactiveSession = ref(session);

  watch(reactiveSession, async (newSession, oldSession) => {
    if (newSession) {
      try {
        token.value = await newSession.getToken();
      } catch (error) {
        console.error("Error getting token:", error);
      }
    }
  }, { immediate: true });

  const friends = ref([])
  const requests = ref([])
  const selectedFriend = ref(null)
  const messages = ref([])
  const newMessage = ref('')
  const requestProfile = ref(null)
  const pusher = ref(null)
  const channel = ref(null)
  const chatLoading = ref(false)
  const friendsLoading = ref(true)
  const notificationPusher = ref(null)
  const notifications = ref([])

  const pusherConfig = {
    appKey: useRuntimeConfig().public.pusherAppKey,
    cluster: "us2",
  }

  const fetchFriends = async () => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${props.user.id}&kind=friend`, {
        method: 'GET',
      })
      if (!response.ok) throw new Error('Failed to fetch friends')
      friends.value = await response.json()
      friendsLoading.value = false
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error fetching friends')
    }
  }

  const fetchNotifications = async () => {
    try{
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${props.user.id}&kind=notifications`, {
        method: 'GET',
      })
      if (!response.ok) throw new Error('Failed to fetch notifications')
      const data = await response.json();
      notifications.value = data ? [...data] : [];
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error fetching notifications')
    }
  }

  const fetchRequests = async () => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${props.user.id}&kind=request`, {
        method: 'GET',
      })
      if (!response.ok) throw new Error('Failed to fetch friend requests')
      requests.value = await response.json()
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error fetching friend requests')
    }
  }

  const selectRequest = (request) => {
    requestProfile.value = request
    selectedFriend.value = null
  }

  const selectFriend = async (friend) => {
    selectedFriend.value = friend
    requestProfile.value = null
    messages.value = []
    await getMessages()
    subscribeToChatChannel()
  }

  const deselectFriend = () => {
    selectedFriend.value = null
    requestProfile.value = null
    messages.value = []
  }

  const acceptRequest = async (request) => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${request.email}&operation=add`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })
      if (!response.ok) throw new Error('Failed to accept friend request')
      friends.value.push(request)
      requests.value = requests.value.filter((r) => r.email !== request.email)
      emit('toast-update', `Successfully connected with ${request.name}`)
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error accepting friend request')
    }
  }

  const denyRequest = async (request) => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${request.email}&operation=remove`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })
      if (!response.ok) throw new Error('Failed to deny friend request')
      requests.value = requests.value.filter((r) => r.email !== request.email)
      emit('toast-update', `${request.name}'s friend request denied`)
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error denying friend request')
    }
  }

  const removeFriend = async (friend) => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${friend.email}&operation=remove`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })
      if (!response.ok) throw new Error('Failed to remove friend')
      friends.value = friends.value.filter((f) => f.email !== friend.email)
      if (selectedFriend.value && selectedFriend.value.email === friend.email) {
        deselectFriend()
      }
      emit('toast-update', `${friend.name} removed from friends list`)
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error removing friend')
    }
  }

  const sendMessage = async () => {
    if (!newMessage.value || !selectedFriend.value) return
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const payload = {
        sender_id: props.user.id,
        receiver_email: selectedFriend.value.email,
        content: newMessage.value,
      }
      messages.value.push({
        id: new Date().getTime(),
        sender: 'me',
        text: newMessage.value,
      })
      newMessage.value = ''
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      })
      if (!response.ok) throw new Error('Failed to send message')
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error sending message')
    }
  }

  const getMessages = async () => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      const response = await fetch(`https://www.pairgrid.com/api/getmessages/getmessages?user_id=${props.user.id}&friend_id=${selectedFriend.value.id}`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })
      if (!response.ok) throw new Error('Failed to fetch messages')
      const data = await response.json()
      messages.value = data.map(message => {
        return {
          id: message.created_at,
          sender: message.sender_id == props.user.id ? 'me' : selectedFriend.value.name,
          text: message.encrypted_content,
        }
      })
      chatLoading.value = false
      if(notifications.value.includes(selectedFriend.value.id)) {
        notifications.value = notifications.value.filter((id) => id !== selectedFriend.value.id)
      }
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Session not found, try again.')
    }
  }

  const subscribeToNotifications = () => {
    notificationPusher.value = new Pusher(pusherConfig.appKey, {
      cluster: pusherConfig.cluster,
    })
    const notificationChannel = notificationPusher.value.subscribe(`notifications-${props.user.id}`)
    notificationChannel.bind('new-notification', (data) => {
      if(!notifications.value.includes(data.sender_id) && (!selectedFriend.value || data.sender_id != selectedFriend.value.id))
        notifications.value.push(data.sender_id)
    })
  }

  const subscribeToChatChannel = async () => {
    if (!selectedFriend.value) return
    unsubscribeFromChatChannel()
    const friendID = selectedFriend.value.id
    const firstID = props.user.id < friendID ? props.user.id : friendID
    const secondID = props.user.id > friendID ? props.user.id : friendID
    const newChannel = `private-chat-${firstID}-${secondID}`
    pusher.value = new Pusher(pusherConfig.appKey, {
      cluster: pusherConfig.cluster,
      authEndpoint: 'https://www.pairgrid.com/api/pusherauth/pusherauth',
      auth: {
        headers: {
          'Accept':'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
      },
    })
    channel.value = pusher.value.subscribe(newChannel)
    channel.value.bind('new-message', (data) => {
      if (data.sender_id != props.user.id) {
        messages.value.push({
          id: data.created_at,
          sender: data.sender_id == props.user.id ? 'me' : selectedFriend.value.name,
          text: data.encrypted_content,
        })
        setTimeout(fetch(`https://www.pairgrid.com/api/getmessages/getmessages?user_id=${props.user.id}&friend_id=${selectedFriend.value.id}&notification_stopper=true`, {
          method: 'GET',
        }), 2000)
      }
    })
    pusher.value.connection.bind('error', (err) => {
      console.error('Pusher connection error:', err);
      emit('toast-update', 'Session not found, try again.');
    });
  }

  const unsubscribeFromChatChannel = () => {
    if (channel.value) {
      channel.value.unbind_all()
      channel.value.unsubscribe()
    }
    if (pusher.value) pusher.value.disconnect()
  }

  const unsubscribeFromNotifications = () => {
    if (notificationPusher.value) notificationPusher.value.disconnect();
  }
  const callFriend = async () => {
    try {
      if(!token.value) {
        console.error("Token not available");
        return;
      }
      emit('call-user', selectedFriend.value.name, selectedFriend.value.id)
      const payload = {
        caller_id: props.user.id,
        callee_id: selectedFriend.value.id,
        type: "voice",
        caller_name: props.user.fullName,
      }
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify(payload),
      })
      if (!response.ok) throw new Error('Failed to call user')
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error calling user')
    }
  }
  const shareScreen = () => {
    emit('toast-update', 'Feature coming soon')
  }

  onMounted(() => {
    subscribeToNotifications()
    fetchFriends()
    fetchRequests()
    fetchNotifications()
  })

  onBeforeUnmount(() => {
    unsubscribeFromChatChannel()
    unsubscribeFromNotifications()
  })
</script>