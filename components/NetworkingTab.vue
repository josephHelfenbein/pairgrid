<template>
    <div>
      <h2 class="text-xl md:text-2xl text-center font-bold mb-4">Recommended Connections</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-if="loading" class="flex justify-center items-center w-full md:w-screen h-64">
          <Loader size="120px" />
        </div>
        <div v-if="recommendedPeople.length === 0 && !loading" class="flex justify-center items-center h-64 w-full md:w-screen">
          <p class="text-xs text-center text-gray-500">No recommended connections found</p>
        </div>
        <Card v-for="person in recommendedPeople">
          <div class="ml-4 flex items-center">
            <img :src="person.profile_picture" class="w-16 h-16 rounded-full object-cover" />
            <CardHeader class="pl-4">
              <CardTitle>{{ person.name }}</CardTitle>
              <CardDescription>{{ person.specialty + ', ' + person.occupation }}</CardDescription>
            </CardHeader>
          </div>
          <CardContent class="flex flex-col justify-between">
            <div>
              <p class="mb-2">{{ person.bio }}</p>
              <div class="flex flex-wrap space-x-2 text-sm mb-1">
                <p class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2" v-for="language in person.language">{{ language }}</p>
              </div>
              <div class="flex flex-wrap space-x-2 text-sm mb-3">
                <p class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2" v-for="interest in person.interests">{{ interest }}</p>
              </div>
            </div>
            <Button v-if="!sentTo.includes(person)" class="outline outline-2 outline-violet-600 bg-violet-900" @click="connect(person)">Connect</Button>
            <Button v-else disabled class="bg-gray-500 cursor-not-allowed">Request Sent</Button>
          </CardContent>
        </Card>
      </div>
      <div class="flex justify-center mt-16">
        <Button v-if="!loading && currentPage>1" class="outline outline-2 outline-violet-600 bg-violet-900" @click="fetchRecommendedPeople(currentPage-1)">Previous</Button>
        <p v-if="recommendedPeople.length===10 || currentPage>1" class="mx-4">Page {{currentPage}}</p>
        <Button v-if="!loading && recommendedPeople.length === 10" class="outline outline-2 outline-violet-600 bg-violet-900" @click="fetchRecommendedPeople(currentPage+1)">Next</Button>
      </div>
    </div>
  </template>
  
  <script setup>
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
  import { defineProps, defineEmits, onMounted, ref } from 'vue';
  import Loader from '@/components/Loader.vue';
  import { useSession } from '@clerk/vue'


  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })
  const user = props.user;
  const emit = defineEmits(['toast-update']);
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

  const recommendedPeople = ref([]);
  const sentTo = ref([]);
  const error = ref(null);
  const loading = ref(true);
  const currentPage = ref(1);
  const fetchRecommendedPeople = async (page = 1) =>{
    loading.value = true;
    try{
      const limit = 10;
      const offset = (page - 1) * limit;
      const response = await fetch(`https://www.pairgrid.com/api/getusers/getusers?user_id=${user.id}&limit=${limit}&offset=${offset}`, {
        method: 'GET',
      });
      if(!response.ok) throw new Error('Failed to fetch recommended people');
      const data = await response.json();
      recommendedPeople.value = data;
      currentPage.value = page;
    } catch (err) {
      console.error(err);
      error.value = err.message;
      emit('toast-update', 'Error fetching recommended connections');
    } finally {
      loading.value = false;
    }
  };
  onMounted(() => {
    fetchRecommendedPeople();
  });
  
  const connect = async (person) => {
    try{
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${person.email}&operation=add`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })
      if(!response.ok) throw new Error('Failed to connect with the user');
      const data = await response.json();
      sentTo.value.push(person);
      emit('toast-update', `Sent friend request to ${person.name}`);
    } catch(err) {
      console.error(err);
      emit('toast-update', 'Error connecting with the user');
    }
  }
  </script>