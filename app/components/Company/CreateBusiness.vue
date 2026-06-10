<script setup lang="ts">
import { Check, Circle, Dot, Plus, Trash2 } from 'lucide-vue-next'
import { toTypedSchema } from '@vee-validate/zod'
import * as z from 'zod'
import { h, ref } from 'vue'
import { Stepper, StepperDescription, StepperItem, StepperSeparator, StepperTitle, StepperTrigger } from '@/components/ui/stepper'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Breadcrumb,
  BreadcrumbItem, BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from "~/components/ui/breadcrumb";
import { toast } from 'vue-sonner'

// Arrays to store multiple accounts
const bankAccounts = ref([{ id: 1, bankName: '', accountNumber: '', accountHolder: '' }])
const momoAccounts = ref([{ id: 1, networkName: '', momoNumber: '', momoAccountHolder: '' }])

const addBankAccount = () => {
  bankAccounts.value.push({
    id: Date.now(),
    bankName: '',
    accountNumber: '',
    accountHolder: ''
  })
}

const removeBankAccount = (id: number) => {
  if (bankAccounts.value.length > 1) {
    bankAccounts.value = bankAccounts.value.filter(acc => acc.id !== id)
  }
}

const addMomoAccount = () => {
  momoAccounts.value.push({
    id: Date.now(),
    networkName: '',
    momoNumber: '',
    momoAccountHolder: ''
  })
}

const removeMomoAccount = (id: number) => {
  if (momoAccounts.value.length > 1) {
    momoAccounts.value = momoAccounts.value.filter(acc => acc.id !== id)
  }
}

const formSchema = [
  z.object({
    fullName: z.string().min(1, 'Full name is required'),
    email: z.string().email('Valid email is required'),
    phoneNumber: z.string().min(1, 'Phone number is required'),
    password: z.string().min(6, 'Password must be at least 6 characters'),
    confirmPassword: z.string(),
  }).refine(
      (values) => values.password === values.confirmPassword,
      {
        message: 'Passwords must match!',
        path: ['confirmPassword'],
      },
  ),
  z.object({
    age: z.string().min(1, 'Age is required'),
    employmentStatus: z.string().min(1, 'Employment status is required'),
    monthlyIncome: z.string().min(1, 'Monthly income is required'),
    dependents: z.string().min(1, 'Number of dependents is required'),
    maritalStatus: z.string().min(1, 'Marital status is required'),
    educationBackground: z.string().min(1, 'Education background is required'),
    homeOwnership: z.string().min(1, 'Home ownership status is required'),
  }),
  z.object({
    dateOfBirth: z.string().min(1, 'Date of birth is required'),
    nationality: z.string().min(1, 'Nationality is required'),
    sex: z.string().min(1, 'Sex is required'),
    idType: z.string().min(1, 'ID type is required'),
    idNumber: z.string().min(1, 'ID number is required'),
    height: z.string().min(1, 'Height is required'),
    placeOfIssuance: z.string().min(1, 'Place of issuance is required'),
    expiryDate: z.string().min(1, 'Expiry date is required'),
    tin: z.string().optional(),
  }),
  z.object({}).passthrough(),  // Allow any fields for dynamic bank accounts
  z.object({}).passthrough(),  // Allow any fields for dynamic momo accounts
  z.object({
    submit: z.string().min(1, 'Please confirm submission'),
  }),
]

const stepIndex = ref(1)
const steps = [
  {
    step: 1,
    title: 'Your details',
    description: 'Provide your name and email',
  },
  {
    step: 2,
    title: 'Personal Information',
    description: 'Enter your personal information',
  },
  {
    step: 3,
    title: 'KYC Information',
    description: 'Enter your kyc information',
  },
  {
    step: 4,
    title: 'Bank Account',
    description: 'Add bank accounts',
  },
  {
    step: 5,
    title: 'Momo Account',
    description: 'Add momo accounts',
  },
  {
    step: 6,
    title: 'Review',
    description: 'Review and Confirm submission',
  },
]

function onSubmit(values: any) {
  const submissionData = {
    ...values,
    bankAccounts: bankAccounts.value,
    momoAccounts: momoAccounts.value
  }
  
  toast.promise(
    () =>
      new Promise(resolve =>
        setTimeout(() => resolve({ name: values.compnayName }), 2000),
      ),
    {
      loading: 'Submitting application...',
      success: (data: { name: string }) => `Application for ${data.name} has been submitted successfully!`,
      error: 'Error submitting application',
    },
  )
}
</script>

<template>
  <div class="flex min-h-screen w-full flex-col bg-muted/40">
  <Form class="md:py-6"
      v-slot="{ meta, values, validate }"
      as="" keep-values :validation-schema="toTypedSchema(formSchema[stepIndex - 1])"
  >
    <Stepper v-slot="{ isNextDisabled, isPrevDisabled, nextStep, prevStep }" v-model="stepIndex" class="block w-full max-w-screen-8xl md:py-16 md:px-28">
      <form
          @submit="(e) => {
          e.preventDefault()
          validate()

          if (stepIndex === steps.length && meta.valid) {
            onSubmit(values)
          }
        }"
      >
        <div class="flex w-full flex-start gap-2 mb-8">
          <StepperItem
              v-for="step in steps"
              :key="step.step"
              v-slot="{ state }"
              class="relative flex w-full flex-col items-center justify-center"
              :step="step.step"
          >
            <StepperSeparator
                v-if="step.step !== steps[steps.length - 1].step"
                class="absolute left-[calc(50%+20px)] right-[calc(-50%+10px)] top-5 block h-0.5 shrink-0 rounded-full bg-muted group-data-[state=completed]:bg-primary transition-all duration-300"
            />

            <StepperTrigger as-child>
              <Button
                  :variant="state === 'completed' || state === 'active' ? 'default' : 'outline'"
                  size="icon"
                  class="z-10 rounded-full shrink-0 transition-all duration-300"
                  :class="[state === 'active' && 'ring-2 ring-ring ring-offset-2 ring-offset-background']"
                  :disabled="state !== 'completed' && !meta.valid"
              >
                <Check v-if="state === 'completed'" class="size-5" />
                <Circle v-if="state === 'active'" />
                <Dot v-if="state === 'inactive'" />
              </Button>
            </StepperTrigger>

            <div class="mt-5 flex flex-col items-center text-center">
              <StepperTitle
                  :class="[state === 'active' && 'text-primary']"
                  class="text-sm font-semibold transition-all duration-300 lg:text-base"
              >
                {{ step.title }}
              </StepperTitle>
              <StepperDescription
                  :class="[state === 'active' && 'text-primary']"
                  class="sr-only text-xs text-muted-foreground transition-all duration-300 md:not-sr-only lg:text-sm"
              >
                {{ step.description }}
              </StepperDescription>
            </div>
          </StepperItem>
        </div>

        <div class="min-h-[400px]">
          <div class="flex flex-col gap-4">
            <template v-if="stepIndex === 1">
              <FormField v-slot="{ componentField }" name="fullName">
                <FormItem>
                  <FormLabel>Full Name</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter your full name" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="email">
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input type="email" placeholder="Enter your email" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="phoneNumber">
                <FormItem>
                  <FormLabel>Phone Number</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter your phone number" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="password">
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="Enter password" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="confirmPassword">
                <FormItem>
                  <FormLabel>Confirm Password</FormLabel>
                  <FormControl>
                    <Input type="password" placeholder="Confirm password" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </template>

            <template v-if="stepIndex === 2">
              <FormField v-slot="{ componentField }" name="age">
                <FormItem>
                  <FormLabel>Age</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter your age" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="employmentStatus">
                <FormItem>
                  <FormLabel>Employment Status</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your employment status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="employed">Employed</SelectItem>
                        <SelectItem value="self-employed">Self-Employed</SelectItem>
                        <SelectItem value="student">Student</SelectItem>
                        <SelectItem value="unemployed">Unemployed</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="monthlyIncome">
                <FormItem>
                  <FormLabel>Monthly Income</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter monthly income" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="dependents">
                <FormItem>
                  <FormLabel>Number of Dependents</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter number of dependents" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="maritalStatus">
                <FormItem>
                  <FormLabel>Marital Status</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your marital status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="single">Single</SelectItem>
                        <SelectItem value="married">Married</SelectItem>
                        <SelectItem value="divorced">Divorced</SelectItem>
                        <SelectItem value="widow">Widow</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="educationBackground">
                <FormItem>
                  <FormLabel>Education Background</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your educational background" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="high-school">High School</SelectItem>
                        <SelectItem value="bachelor">Bachelor's Degree</SelectItem>
                        <SelectItem value="master">Master's Degree</SelectItem>
                        <SelectItem value="phd">PhD</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="homeOwnership">
                <FormItem>
                  <FormLabel>Home Ownership Status</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your status" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="owned">Owned</SelectItem>
                        <SelectItem value="rented">Rented</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>
            </template>

            <template v-if="stepIndex === 3">
              <FormField v-slot="{ componentField }" name="dateOfBirth">
                <FormItem>
                  <FormLabel>Date of Birth</FormLabel>
                  <FormControl>
                    <Input type="date" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="nationality">
                <FormItem>
                  <FormLabel>Nationality</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your nationality" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="ghanaian">Ghanaian</SelectItem>
                        <SelectItem value="nigerian">Nigerian</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="sex">
                <FormItem>
                  <FormLabel>Sex</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select your sex" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="male">Male</SelectItem>
                        <SelectItem value="female">Female</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="idType">
                <FormItem>
                  <FormLabel>ID Type</FormLabel>
                  <Select v-bind="componentField">
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select ID type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="ghana-card">Ghana Card</SelectItem>
                        <SelectItem value="passport">Passport</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="idNumber">
                <FormItem>
                  <FormLabel>ID Number</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter ID number" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="height">
                <FormItem>
                  <FormLabel>Height (cm)</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter height" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="placeOfIssuance">
                <FormItem>
                  <FormLabel>Place of Issuance</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter place of issuance" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="expiryDate">
                <FormItem>
                  <FormLabel>Expiry Date</FormLabel>
                  <FormControl>
                    <Input type="date" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="tin">
                <FormItem>
                  <FormLabel>TIN (Optional)</FormLabel>
                  <FormControl>
                    <Input type="text" placeholder="Enter TIN" v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </template>

            <template v-if="stepIndex === 4">
              <div class="space-y-6">
                <div v-for="(account, index) in bankAccounts" :key="account.id" class="relative">
                  <Card>
                    <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-4">
                      <CardTitle class="text-base">Bank Account {{ index + 1 }}</CardTitle>
                      <Button 
                        v-if="bankAccounts.length > 1"
                        type="button"
                        variant="ghost" 
                        size="icon"
                        @click="removeBankAccount(account.id)"
                        class="h-8 w-8"
                      >
                        <Trash2 class="h-4 w-4 text-destructive" />
                      </Button>
                    </CardHeader>
                    <CardContent class="space-y-4">
                      <div>
                        <label class="text-sm font-medium">Bank Name</label>
                        <Select v-model="account.bankName">
                          <SelectTrigger>
                            <SelectValue placeholder="Select Bank" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectGroup>
                              <SelectItem value="stanbic-bank">Stanbic Bank</SelectItem>
                              <SelectItem value="fidelity-bank">Fidelity Bank</SelectItem>
                              <SelectItem value="gcb-bank">GCB Bank</SelectItem>
                              <SelectItem value="absa-bank">Absa Bank</SelectItem>
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                      </div>

                      <div>
                        <label class="text-sm font-medium">Account Number</label>
                        <Input 
                          type="text" 
                          v-model="account.accountNumber"
                          placeholder="Enter account number" 
                        />
                      </div>

                      <div>
                        <label class="text-sm font-medium">Account Holder</label>
                        <Input 
                          type="text" 
                          v-model="account.accountHolder"
                          placeholder="Enter account holder name" 
                        />
                      </div>
                    </CardContent>
                  </Card>
                </div>

                <Button 
                  type="button"
                  variant="outline" 
                  class="w-full"
                  @click="addBankAccount"
                >
                  <Plus class="h-4 w-4 mr-2" />
                  Add Another Bank Account
                </Button>
              </div>
            </template>

            <template v-if="stepIndex === 5">
              <div class="space-y-6">
                <div v-for="(account, index) in momoAccounts" :key="account.id" class="relative">
                  <Card>
                    <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-4">
                      <CardTitle class="text-base">Mobile Money Account {{ index + 1 }}</CardTitle>
                      <Button 
                        v-if="momoAccounts.length > 1"
                        type="button"
                        variant="ghost" 
                        size="icon"
                        @click="removeMomoAccount(account.id)"
                        class="h-8 w-8"
                      >
                        <Trash2 class="h-4 w-4 text-destructive" />
                      </Button>
                    </CardHeader>
                    <CardContent class="space-y-4">
                      <div>
                        <label class="text-sm font-medium">Network Name</label>
                        <Select v-model="account.networkName">
                          <SelectTrigger>
                            <SelectValue placeholder="Select network" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectGroup>
                              <SelectItem value="mtn">MTN</SelectItem>
                              <SelectItem value="telecel">Telecel</SelectItem>
                              <SelectItem value="airteltigo">AirtelTigo</SelectItem>
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                      </div>

                      <div>
                        <label class="text-sm font-medium">Mobile Money Number</label>
                        <Input 
                          type="text" 
                          v-model="account.momoNumber"
                          placeholder="Enter mobile money number" 
                        />
                      </div>

                      <div>
                        <label class="text-sm font-medium">Account Holder</label>
                        <Input 
                          type="text" 
                          v-model="account.momoAccountHolder"
                          placeholder="Enter account holder name" 
                        />
                      </div>
                    </CardContent>
                  </Card>
                </div>

                <Button 
                  type="button"
                  variant="outline" 
                  class="w-full"
                  @click="addMomoAccount"
                >
                  <Plus class="h-4 w-4 mr-2" />
                  Add Another Mobile Money Account
                </Button>
              </div>
            </template>

            <template v-if="stepIndex === 6">
              <div class="space-y-6">
                <div class="rounded-lg border p-6 bg-muted/50">
                  <h3 class="text-lg font-semibold mb-4">Review Your Information</h3>
                  <p class="text-sm text-muted-foreground mb-6">
                    Please review all the information you've provided before submitting.
                  </p>

                  <div class="space-y-6">
                    <!-- Personal Details -->
                    <Card>
                      <CardHeader>
                        <CardTitle class="text-base">Personal Details</CardTitle>
                      </CardHeader>
                      <CardContent class="space-y-2 text-sm">
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Full Name:</span>
                          <span class="font-medium">{{ values.fullName || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Email:</span>
                          <span class="font-medium">{{ values.email || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Phone Number:</span>
                          <span class="font-medium">{{ values.phoneNumber || 'Not provided' }}</span>
                        </div>
                      </CardContent>
                    </Card>

                    <!-- Personal Information -->
                    <Card>
                      <CardHeader>
                        <CardTitle class="text-base">Personal Information</CardTitle>
                      </CardHeader>
                      <CardContent class="space-y-2 text-sm">
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Age:</span>
                          <span class="font-medium">{{ values.age || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Employment Status:</span>
                          <span class="font-medium">{{ values.employmentStatus || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Monthly Income:</span>
                          <span class="font-medium">{{ values.monthlyIncome || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Dependents:</span>
                          <span class="font-medium">{{ values.dependents || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Marital Status:</span>
                          <span class="font-medium">{{ values.maritalStatus || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Education:</span>
                          <span class="font-medium">{{ values.educationBackground || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Home Ownership:</span>
                          <span class="font-medium">{{ values.homeOwnership || 'Not provided' }}</span>
                        </div>
                      </CardContent>
                    </Card>

                    <!-- KYC Information -->
                    <Card>
                      <CardHeader>
                        <CardTitle class="text-base">KYC Information</CardTitle>
                      </CardHeader>
                      <CardContent class="space-y-2 text-sm">
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Date of Birth:</span>
                          <span class="font-medium">{{ values.dateOfBirth || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Nationality:</span>
                          <span class="font-medium">{{ values.nationality || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Sex:</span>
                          <span class="font-medium">{{ values.sex || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">ID Type:</span>
                          <span class="font-medium">{{ values.idType || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">ID Number:</span>
                          <span class="font-medium">{{ values.idNumber || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Height:</span>
                          <span class="font-medium">{{ values.height ? values.height + ' cm' : 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Place of Issuance:</span>
                          <span class="font-medium">{{ values.placeOfIssuance || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between">
                          <span class="text-muted-foreground">Expiry Date:</span>
                          <span class="font-medium">{{ values.expiryDate || 'Not provided' }}</span>
                        </div>
                        <div class="flex justify-between" v-if="values.tin">
                          <span class="text-muted-foreground">TIN:</span>
                          <span class="font-medium">{{ values.tin }}</span>
                        </div>
                      </CardContent>
                    </Card>

                    <!-- Bank Accounts -->
                    <Card>
                      <CardHeader>
                        <CardTitle class="text-base">Bank Accounts ({{ bankAccounts.length }})</CardTitle>
                      </CardHeader>
                      <CardContent class="space-y-4">
                        <div v-for="(account, index) in bankAccounts" :key="account.id" class="p-4 border rounded-lg space-y-2 text-sm">
                          <div class="font-semibold text-primary mb-2">Account {{ index + 1 }}</div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Bank Name:</span>
                            <span class="font-medium">{{ account.bankName || 'Not provided' }}</span>
                          </div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Account Number:</span>
                            <span class="font-medium">{{ account.accountNumber || 'Not provided' }}</span>
                          </div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Account Holder:</span>
                            <span class="font-medium">{{ account.accountHolder || 'Not provided' }}</span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>

                    <!-- Mobile Money Accounts -->
                    <Card>
                      <CardHeader>
                        <CardTitle class="text-base">Mobile Money Accounts ({{ momoAccounts.length }})</CardTitle>
                      </CardHeader>
                      <CardContent class="space-y-4">
                        <div v-for="(account, index) in momoAccounts" :key="account.id" class="p-4 border rounded-lg space-y-2 text-sm">
                          <div class="font-semibold text-primary mb-2">Account {{ index + 1 }}</div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Network:</span>
                            <span class="font-medium">{{ account.networkName || 'Not provided' }}</span>
                          </div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Mobile Number:</span>
                            <span class="font-medium">{{ account.momoNumber || 'Not provided' }}</span>
                          </div>
                          <div class="flex justify-between">
                            <span class="text-muted-foreground">Account Holder:</span>
                            <span class="font-medium">{{ account.momoAccountHolder || 'Not provided' }}</span>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  </div>
                </div>

                <FormField v-slot="{ componentField }" name="submit">
                  <FormItem>
                    <FormLabel>Confirm Submission</FormLabel>
                    <Select v-bind="componentField">
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select answer" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="yes">Yes, submit my application</SelectItem>
                          <SelectItem value="no">No, I need to review</SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                </FormField>
              </div>
            </template>
          </div>
        </div>

        <div class="flex items-center justify-between mt-8 pt-4 border-t">
          <Button :disabled="isPrevDisabled" variant="outline" size="sm" @click="prevStep()">
            Back
          </Button>
          <div class="flex items-center gap-3">
            <Button v-if="stepIndex !== 6" :type="meta.valid ? 'button' : 'submit'" :disabled="isNextDisabled" size="sm" @click="meta.valid && nextStep()">
              Next
            </Button>
            <Button
                v-if="stepIndex === 6" size="sm" type="submit"
            >
              Submit
            </Button>
          </div>
        </div>
      </form>
    </Stepper>
  </Form>

    </div>
</template>