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
              :friendsLoading="friendsLoading"
              :selectedFriend="selectedFriend"
              @selectFriend="selectFriend"
              @acceptRequest="acceptRequest"
              @denyRequest="denyRequest"
            />
          </ScrollArea>
        </CardContent>
      </Card>
    </div>

    <div class="md:hidden w-full">
      <div v-if="!selectedFriend" class="h-full">
        <Card class="h-full">
          <CardHeader>
            <CardTitle>Friends</CardTitle>
          </CardHeader>
          <CardContent>
            <ScrollArea class="h-[calc(100vh-300px)]">
              <FriendsList 
                :friends="friends" 
                :requests="requests" 
                :friendsLoading="friendsLoading"
                :selectedFriend="selectedFriend"
                @selectFriend="selectFriend"
                @acceptRequest="acceptRequest"
                @denyRequest="denyRequest"
              />
            </ScrollArea>
          </CardContent>
        </Card>
      </div>

      <div v-else class="h-full">
        <Card class="h-full">
          <CardHeader class="flex flex-row justify-between items-center">
            <button @click="deselectFriend" class="p-2">
              <ChevronLeft class="h-6 w-6" />
            </button>
            <CardTitle class="flex-shrink-0 flex items-center">
              {{ selectedFriend.name }}
            </CardTitle>
            <FriendOptions 
              :selectedFriend="selectedFriend"
              :friendProfile="friendProfile"
              @removeFriend="removeFriend"
            />
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
    </div>

    <Card class="hidden md:block md:w-2/3">
      <CardHeader class="flex flex-row justify-between items-center">
        <CardTitle class="flex-shrink-0 flex items-center">
          {{ selectedFriend ? selectedFriend.name : 'Select a friend' }}
        </CardTitle>
        <FriendOptions 
          v-if="selectedFriend"
          :selectedFriend="selectedFriend"
          :friendProfile="friendProfile"
          @removeFriend="removeFriend"
        />
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
        <div v-else class="flex justify-center items-center h-full">
          <p class="text-gray-500">Select a friend to start chatting</p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
  
<script setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue'
  import { ChevronLeft } from 'lucide-vue-next'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { ScrollArea } from '@/components/ui/scroll-area'
  import { useRuntimeConfig } from '#app'
  import CryptoJS from 'crypto-js'
  import Pusher from 'pusher-js'

  import FriendsList from './FriendsList.vue'
  import FriendOptions from './FriendOptions.vue'
  import ChatArea from './ChatArea.vue'

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })

  const emit = defineEmits(['toast-update'])

  const friends = ref([])
  const requests = ref([])
  const selectedFriend = ref(null)
  const messages = ref([])
  const newMessage = ref('')
  const friendProfile = ref(null)
  const pusher = ref(null)
  const channel = ref(null)
  const chatLoading = ref(false)
  const friendsLoading = ref(true)

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

  const selectFriend = (friend) => {
    selectedFriend.value = friend
    messages.value = []
    fetchFriendProfile(friend)
  }

  const deselectFriend = () => {
    selectedFriend.value = null
    messages.value = []
  }

  const acceptRequest = async (request) => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${request.email}&operation=add`, {
        method: 'GET',
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
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${request.email}&operation=remove`, {
        method: 'GET',
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
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${props.user.id}&friend_email=${friend.email}&operation=remove`, {
        method: 'GET',
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
      const encryptionKey = generateEncryptionKey(props.user.id)
      const encryptedMessage = encryptMessage(newMessage.value, encryptionKey)
      const payload = {
        sender_id: props.user.id,
        receiver_email: selectedFriend.value.email,
        content: encryptedMessage.encryptedData,
        key: encryptedMessage.iv,
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
        },
        body: JSON.stringify(payload),
      })
      if (!response.ok) throw new Error('Failed to send message')
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error sending message')
    }
  }

  const generateEncryptionKey = (userID) => {
    const serverSideSecret = useRuntimeConfig().public.encryptionKey
    return CryptoJS.SHA256(userID + serverSideSecret)
  }

  const encryptMessage = (message, key) => {
    const iv = CryptoJS.lib.WordArray.random(16)
    const encrypted = CryptoJS.AES.encrypt(message, key, { iv: iv })
    
    return {
      encryptedData: encrypted.ciphertext.toString(CryptoJS.enc.Hex),
      iv: iv.toString(CryptoJS.enc.Hex)
    }
  }

  const decryptMessage = (encryptedMessage, key, iv) => {
    const ivWordArray = CryptoJS.enc.Hex.parse(iv)
    const encryptedWordArray = CryptoJS.enc.Hex.parse(encryptedMessage)

    const decrypted = CryptoJS.AES.decrypt(
      { ciphertext: encryptedWordArray },
      key,
      { iv: ivWordArray }
    )

    return decrypted.toString(CryptoJS.enc.Utf8)
  }

  const fetchFriendProfile = async (friend) => {
    try {
      chatLoading.value = true
      const emailData = { email: friend.email }
      const response = await fetch('https://www.pairgrid.com/api/getuser/getuser', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(emailData),
      })
      if (!response.ok) throw new Error('Failed to fetch user profile')
      friendProfile.value = await response.json()
      await getMessages()
      subscribeToChatChannel()
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error fetching friend profile')
    }
  }

  const getMessages = async () => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/getmessages/getmessages?user_id=${props.user.id}&friend_id=${friendProfile.value.id}`, {
        method: 'GET',
      })
      if (!response.ok) throw new Error('Failed to fetch messages')
      const data = await response.json()
      messages.value = data.map(message => {
        const decryptedMessage = decryptMessage(
          message.encrypted_content, 
          generateEncryptionKey(message.sender_id), 
          message.key
        )

        return {
          id: message.created_at,
          sender: message.sender_id == props.user.id ? 'me' : selectedFriend.value.name,
          text: decryptedMessage,
        }
      })
      chatLoading.value = false
    } catch (err) {
      console.error(err)
      emit('toast-update', 'Error loading chat')
    }
  }

  const subscribeToChatChannel = () => {
    if (!selectedFriend.value || !friendProfile.value) return
    unsubscribeFromChatChannel()
    const friendID = friendProfile.value.id
    const firstID = props.user.id < friendID ? props.user.id : friendID
    const secondID = props.user.id > friendID ? props.user.id : friendID
    const newChannel = `chat-${firstID}-${secondID}`
    pusher.value = new Pusher(pusherConfig.appKey, {
      cluster: pusherConfig.cluster,
    })
    channel.value = pusher.value.subscribe(newChannel)
    channel.value.bind('new-message', (data) => {
      const decryptedMessage = decryptMessage(data.encrypted_content, generateEncryptionKey(data.sender_id), data.key)
      
      if (data.sender_id != props.user.id) {
        messages.value.push({
          id: data.created_at,
          sender: data.sender_id == props.user.id ? 'me' : selectedFriend.value.name,
          text: decryptedMessage,
        })
      }
    })
  }

  const unsubscribeFromChatChannel = () => {
    if (channel.value) {
      channel.value.unbind_all()
      channel.value.unsubscribe()
    }
    if (pusher.value) {
      pusher.value.disconnect()
    }
  }

  onMounted(() => {
    fetchFriends()
    fetchRequests()
  })

  onBeforeUnmount(() => {
    unsubscribeFromChatChannel()
  })
</script>