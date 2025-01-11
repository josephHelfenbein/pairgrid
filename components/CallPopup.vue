<template>
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
</template>
<script setup>
    import { ref } from 'vue';

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
    export function triggerIncomingCall(name){
        callerName.value = name;
        showCallPopup.value = true;
    }
</script>