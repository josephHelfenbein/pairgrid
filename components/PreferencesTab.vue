<template>
    <Card v-if="preferences">
      <CardHeader>
        <CardTitle>Preferences</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit" class="space-y-6">
          <FormField v-slot="{componentField}" name="bio">
            <FormItem>
              <FormLabel>Bio</FormLabel>
              <FormControl>
                <Input 
                type="text" 
                placeholder="Tell us about yourself" 
                v-bind="componentField" 
                v-model="preferences.bio"/>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
          

          <div class="space-y-2">
            <Label>Specialty (Select one)</Label>
            <div class="space-y-2">
              <div v-for="specialty in specialties" :key="specialty" class="flex items-center space-x-2">
                <Checkbox
                  :id="specialty"
                  :checked="preferences.specialty==specialty"
                  @update:checked="toggleSpecialty(specialty)"
                />
                <Label :for="specialty">{{ specialty }}</Label>
              </div>
            </div>
          </div>

          <FormField v-slot="{componentField}" name="occupation">
            <FormItem>
              <FormLabel>Occupation</FormLabel>
              <FormControl>
                <Select v-bind="componentField" v-model="preferences.occupation">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="occupation in occupations" :key="occupation" :value="occupation">
                      {{ occupation }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <div class="space-y-2">
            <Label>Interests (Select multiple)</Label>
            <div class="space-y-2">
              <div v-for="interest in interests" :key="interest" class="flex items-center space-x-2">
                <Checkbox
                  :id="interest"
                  :checked="preferences.interests.includes(interest)"
                  @update:checked="toggleInterest(interest)"
                />
                <Label :for="interest">{{ interest }}</Label>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <Label>Programming languages (Select multiple)</Label>
            <div class="space-y-2">
              <div v-for="language in languages" :key="language" class="flex items-center space-x-2">
                <Checkbox
                  :id="language"
                  :checked="preferences.language.includes(language)"
                  @update:checked="toggleLanguage(language)"
                />
                <Label :for="language">{{ language }}</Label>
              </div>
            </div>
          </div>
          <Button type="submit">Save Preferences</Button>
        </form>
      </CardContent>
    </Card>
  </template>
  
  <script setup>
  import { defineProps, reactive, defineEmits, ref, watch } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { Label } from '@/components/ui/label'
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
  import { toTypedSchema } from '@vee-validate/zod'
  import {useForm} from 'vee-validate'
  import * as z from 'zod'
  import { Checkbox } from '@/components/ui/checkbox'
  import { useSession } from '@clerk/vue'
  
  const props = defineProps({
    preferences: {
      type: Object,
      required: true,
    },
    user: {
      type: Object,
      required: true,
    }
  })
  const emit = defineEmits(['update-preferences']);
  const preferences = reactive({ ...props.preferences });
  const user = props.user;
  const token = ref(null);
  const { session } = useSession();
  watch(() => session, async (newSession) => {
    console.log("Getting token...");
    token.value = await newSession.getToken();
    console.log(token.value);
  })

  const occupations = [
    'Middle School Student',
    'High School Student',
    'Undergraduate Student',
    'Graduate Student',
    'Professional',
    'Hobbyist',
    'Educator',
  ]

  const languages = [
    'JavaScript',
    'TypeScript',
    'Python',
    'Java',
    'Ruby',
    'Go',
    'Dart',
    'C/C++',
    'C#',
    'PHP',
    'Swift',
    'Kotlin',
    'Rust',
    'Scala',
    'Perl',
    'R',
    'Haskell',
    'Lua',
  ]

  const specialties = [
    'Full Stack Developer',
    'Front-End Developer',
    'Back-End Developer',
    'Mobile Developer',
    'Data Scientist',
    'Designer',
    'Product Manager',
    'DevOps Engineer',
    'QA Engineer',
    'Machine Learning Engineer',
    'Embedded Systems Engineer',
    'Game Developer',
    'Cloud Engineer',
  ]

  const interests = [
    'AR/VR',
    'Blockchain',
    'Cybersecurity',
    'IoT',
    'Big Data',
    'Cloud Computing',
    'Web Development',
    'Mobile Development',
    'Machine Learning',
    'Game Development',
    'UI/UX Design',
    'Data Science',
    'DevOps',
    'Low-level Programming',
    'Graphics Programming',
  ]

  const toggleSpecialty = (interest) => {
    if(preferences.specialty==interest) preferences.specialty = '';
    else preferences.specialty = interest;
  }
  const toggleLanguage = (interest) => {
    let index = preferences.language.indexOf(interest);
    if(index==-1) preferences.language.push(interest);
    else preferences.language.splice(index, 1);
  }
  const toggleInterest = (interest) => {
    let index = preferences.interests.indexOf(interest);
    if(index==-1) preferences.interests.push(interest);
    else preferences.interests.splice(index, 1);
  }
  
  const formSchema = toTypedSchema(z.object({
    occupation: z.enum([
      'Middle School Student',
      'High School Student',
      'Undergraduate Student',
      'Graduate Student',
      'Professional',
      'Hobbyist',
      'Educator',
    ], {required_error: 'Please select an occupation'}),
    bio: z.string().min(10).max(250),
  }))
  const { handleSubmit } = useForm({
    validationSchema: formSchema,
    initialValues: props.preferences,
  })
  const onSubmit = handleSubmit((values)=>{
    preferences.bio = values.bio
    preferences.occupation = values.occupation
    const data = {
      id: user.id,
      bio: preferences.bio,
      language: [...preferences.language],
      specialty: preferences.specialty,
      interests: [...preferences.interests],
      occupation: preferences.occupation
    };
    if (!token.value) {
      console.error('Token not available');
      return;
    }
    
    fetch('https://www.pairgrid.com/api/updateuser/updateuser', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token.value}`,
      },
      body: JSON.stringify(data),
    }).then((response)=>{ 
      if(response.ok){
        response.json().then(result=>{
          console.log('Preferences updated successfully');
        }).catch(error=>{
          console.error('Error parsing response:', error);
        });
      } else{
        console.error('Failed to update preferences:', response.statusText);
      }
    }).catch(error=>{
      console.error('Error updating preferences:', error);
    });

    emit('update-preferences', preferences);
  });

  </script>