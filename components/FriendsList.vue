<template>
  <div>
    <div class="flex justify-center items-center h-full" v-if="friendsLoading">
      <Loader size="80px" />
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="request in requests"
        :key="request.email"
        class="w-full justify-between flex items-center"
      >
        <div class="flex items-center gap-2">
          <img :src="request.profile_picture" class="w-8 h-8 rounded-full object-cover" />
          <button @click="$emit('selectRequest', request)" class="bg-none text-left text-white">
            <p>{{ request.name }}</p>
          </button>
        </div>
        <div class="flex gap-2">
          <button @click="$emit('acceptRequest', request)" class="p-2 bg-green-500 text-white rounded-full">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
          <button @click="$emit('denyRequest', request)" class="p-2 bg-red-500 text-white rounded-full">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
      <div v-if="friends.length === 0 && !friendsLoading" class="flex justify-center items-center w-full h-full">
        <p class="text-xs text-center text-gray-500">No friends found. Make friends in the Networking tab!</p>
      </div>
      <Button
        v-for="friend in friends"
        :key="friend.email"
        :variant="selectedFriend?.email === friend.email ? 'secondary' : 'ghost'"
        class="w-full justify-start flex items-center"
        @click="$emit('selectFriend', friend)"
      >
        <img :src="friend.profile_picture" class="w-8 h-8 rounded-full object-cover" />
        <div>
          <p class="text-left">
            {{ friend.name }}
          </p>
          <p class="text-sm text-left text-gray-500">{{ getLastSeenText(friend.last_seen) }}</p>
        </div>
      </Button>
    </div>
  </div>
</template>
  
<script setup>
  import { Button } from '@/components/ui/button'
  import Loader from '@/components/Loader'
  
  defineProps({
    friends: Array,
    requests: Array,
    friendsLoading: Boolean,
    selectedFriend: Object,
  })
  
  const getLastSeenText = (lastSeen) => {
    const now = new Date()
    const lastSeenDate = new Date(lastSeen)
    const diffInSeconds = Math.floor((now - lastSeenDate) / 1000)
  
    if (diffInSeconds < 60) {
      return `Last seen just now`
    }
    const diffInMinutes = Math.floor(diffInSeconds / 60)
    if (diffInMinutes < 60) {
      return `Last seen ${diffInMinutes} minute${diffInMinutes > 1 ? 's' : ''} ago`
    }
    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) {
      return `Last seen ${diffInHours} hour${diffInHours > 1 ? 's' : ''} ago`
    }
    const diffInDays = Math.floor(diffInHours / 24)
    return `Last seen ${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`
  }
</script>