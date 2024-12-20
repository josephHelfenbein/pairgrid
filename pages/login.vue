<script setup lang="js">
    import { Clerk } from '@clerk/clerk-js';
    import { watch } from 'vue';
    import { useRoute } from '#imports';
    const isClient = typeof window !== 'undefined';
    const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;
    const initializeClerk = ()=>{
        if(isClient){
            const clerk = new Clerk(clerkPubKey);
            clerk.load().then(()=>{
                if(clerk.user){
                    document.getElementById('app').innerHTML= `<div id="user-button"></div>`;
                    const userButtonDiv = document.getElementById('user-button');
                    clerk.mountUserButton(userButtonDiv);
                }
                else{
                    document.getElementById('app').innerHTML= `<div id="sign-in"></div>`;
                    const signInDiv = document.getElementById('sign-in');
                    clerk.mountSignIn(signInDiv);
                }
            }).catch(error =>{
                console.error("Error loading Clerk:", error);
            });
        }
    }
    const route = useRoute();
    watch(
        () => route?.fullPath,
        (newPath, oldPath) => {
            initializeClerk();
        },
        { immediate: true }
    );
</script>

<template>
    <div class="w-full mt-10 flex justify-center items-center flex-col">
        <div id="app"></div>
    </div>
</template>