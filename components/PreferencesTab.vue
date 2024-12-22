<template>
    <Card>
      <CardHeader>
        <CardTitle>Preferences</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit="onSubmit" class="space-y-6">
          <FormField v-slot="{componentField}" name="bio">
            <FormItem>
              <FormLabel>Bio</FormLabel>
              <FormControl>
                <Input type="text" placeholder="Tell us about yourself" v-bind="componentField"/>
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
                <Select v-bind="componentField">
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
  import { ref } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { Label } from '@/components/ui/label'
  import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
  import { toTypedSchema } from '@vee-validate/zod'
  import {useForm} from 'vee-validate'
  import * as z from 'zod'
  import { Checkbox } from '@/components/ui/checkbox'
  
  const occupations = [
    'Middle School Student',
    'High School Student',
    'Undergraduate Student',
    'Graduate Student',
    'Professional Developer',
    'Hobbyist Developer',
    'Educator',
  ]

  const languages = [
    'JavaScript',
    'TypeScript',
    'Python',
    'Java',
    'Ruby',
    'Go',
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
    'Full-stack Developer',
    'Front-end Developer',
    'Back-end Developer',
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
  
  const preferences = ref({
    bio: '',
    language: [],
    specialty: '',
    interests: [],
    occupation: '',
  })

  const setBio = (bio) => {
    preferences.value.bio = bio;
  }  
  const setOccupation = (occupation) => {
    preferences.value.occupation = occupation;
  }
  const toggleSpecialty = (interest) => {
    if(preferences.value.specialty==interest) preferences.value.specialty = '';
    else preferences.value.specialty = interest;
  }
  const toggleLanguage = (interest) => {
    let index = preferences.value.language.indexOf(interest);
    if(index==-1) preferences.value.language.push(interest);
    else preferences.value.language.splice(index, 1);
  }
  const toggleInterest = (interest) => {
    let index = preferences.value.interests.indexOf(interest);
    if(index==-1) preferences.value.interests.push(interest);
    else preferences.value.interests.splice(index, 1);
  }
  
  const formSchema = toTypedSchema(z.object({
    occupation: z.enum(['Middle School Student',
    'High School Student',
    'Undergraduate Student',
    'Graduate Student',
    'Professional Developer',
    'Hobbyist Developer',
    'Educator',], {required_error: 'Please select an occupation'}),
    bio: z.string().min(10).max(250),
  }))
  const {handleSubmit} = useForm({validationSchema: formSchema});
  const onSubmit = handleSubmit((values)=>{
    setBio(values.bio);
    setOccupation(values.occupation);
    console.log(preferences)
  });
  </script>