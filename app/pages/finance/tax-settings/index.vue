<script setup lang="ts">
import { Loader2, Plus, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface TaxRate {
  id: string
  name: string
  rate: string
  compound: boolean
  order: number
  active: boolean
}

interface CompanySettings {
  id: string
  companyName: string
  tin?: string | null
  vatNumber?: string | null
  address?: string | null
  phone?: string | null
  email?: string | null
}

const { data, refresh, pending: isLoading } = await useFetch<{ settings: CompanySettings | null, rates: TaxRate[] }>('/api/tax-settings')

const isSavingSettings = ref(false)
const settingsForm = reactive({
  companyName: '',
  tin: '',
  vatNumber: '',
  address: '',
  phone: '',
  email: '',
})

watch(data, (value) => {
  if (value?.settings) {
    settingsForm.companyName = value.settings.companyName ?? ''
    settingsForm.tin = value.settings.tin ?? ''
    settingsForm.vatNumber = value.settings.vatNumber ?? ''
    settingsForm.address = value.settings.address ?? ''
    settingsForm.phone = value.settings.phone ?? ''
    settingsForm.email = value.settings.email ?? ''
  }
}, { immediate: true })

async function saveSettings() {
  isSavingSettings.value = true
  try {
    await $fetch('/api/tax-settings', { method: 'PATCH', body: settingsForm })
    toast.success('Tax settings saved')
    await refresh()
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to save settings')
  }
  finally {
    isSavingSettings.value = false
  }
}

const rates = computed(() => data.value?.rates ?? [])

const isAddingRate = ref(false)
const newRate = reactive({ name: '', rate: 0, compound: false })

async function addRate() {
  if (!newRate.name || newRate.rate < 0) {
    toast.error('Enter a tax name and a valid rate')
    return
  }
  isAddingRate.value = true
  try {
    await $fetch('/api/tax-settings/rates', {
      method: 'POST',
      body: { ...newRate, order: rates.value.length },
    })
    newRate.name = ''
    newRate.rate = 0
    newRate.compound = false
    await refresh()
    toast.success('Tax rate added')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to add tax rate')
  }
  finally {
    isAddingRate.value = false
  }
}

async function updateRate(rate: TaxRate, patch: Partial<Pick<TaxRate, 'name' | 'rate' | 'compound' | 'active'>>) {
  try {
    await $fetch(`/api/tax-settings/rates/${rate.id}`, {
      method: 'PATCH',
      body: patch.rate !== undefined ? { ...patch, rate: Number(patch.rate) } : patch,
    })
    await refresh()
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to update tax rate')
  }
}

async function removeRate(rate: TaxRate) {
  try {
    await $fetch(`/api/tax-settings/rates/${rate.id}`, { method: 'DELETE' })
    await refresh()
    toast.success('Tax rate removed')
  }
  catch {
    toast.error('Failed to remove tax rate')
  }
}
</script>

<template>
  <FinanceLayout>
    <div class="space-y-6">
      <div>
        <h3 class="text-lg font-medium">
          Tax Settings
        </h3>
        <p class="text-sm text-muted-foreground">
          Company tax registration details and the tax rates applied to new invoices and quotations.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Company Tax Registration</CardTitle>
          <CardDescription>
            Shown on invoices and GRA receipts.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="isLoading" class="text-sm text-muted-foreground">
            Loading...
          </div>
          <form v-else class="grid gap-4" @submit.prevent="saveSettings">
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="companyName">Company Name</Label>
                <Input id="companyName" v-model="settingsForm.companyName" />
              </div>
              <div class="grid gap-2">
                <Label for="tin">TIN (Tax Identification Number)</Label>
                <Input id="tin" v-model="settingsForm.tin" placeholder="e.g. C0001234567" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="vatNumber">VAT Registration Number</Label>
                <Input id="vatNumber" v-model="settingsForm.vatNumber" />
              </div>
              <div class="grid gap-2">
                <Label for="phone">Phone</Label>
                <Input id="phone" v-model="settingsForm.phone" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="email">Email</Label>
                <Input id="email" v-model="settingsForm.email" type="email" />
              </div>
              <div class="grid gap-2">
                <Label for="address">Address</Label>
                <Input id="address" v-model="settingsForm.address" />
              </div>
            </div>
            <Button type="submit" class="justify-self-end" :disabled="isSavingSettings">
              <Loader2 v-if="isSavingSettings" class="mr-2 h-4 w-4 animate-spin" />
              Save Settings
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Tax Rates</CardTitle>
          <CardDescription>
            Levies (NHIL, GETFund, COVID-19 Levy) are calculated on the invoice subtotal. Compound taxes (VAT) are calculated on the subtotal plus those levies, matching GRA's standard VAT scheme.
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Rate (%)</TableHead>
                  <TableHead>Compound</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead class="text-right">
                    Actions
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="rate in rates" :key="rate.id">
                  <TableCell>
                    <Input
                      :model-value="rate.name"
                      class="h-8"
                      @change="(e: Event) => updateRate(rate, { name: (e.target as HTMLInputElement).value })"
                    />
                  </TableCell>
                  <TableCell>
                    <Input
                      :model-value="rate.rate"
                      type="number" min="0" step="0.01" class="h-8 w-24"
                      @change="(e: Event) => updateRate(rate, { rate: Number((e.target as HTMLInputElement).value) } as any)"
                    />
                  </TableCell>
                  <TableCell>
                    <Checkbox :model-value="rate.compound" @update:model-value="(v: boolean) => updateRate(rate, { compound: v })" />
                  </TableCell>
                  <TableCell>
                    <Checkbox :model-value="rate.active" @update:model-value="(v: boolean) => updateRate(rate, { active: v })" />
                  </TableCell>
                  <TableCell class="text-right">
                    <Button size="icon" variant="ghost" @click="removeRate(rate)">
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
                <TableRow v-if="rates.length === 0">
                  <TableCell colspan="5" class="text-center text-muted-foreground">
                    No tax rates configured yet.
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <div class="grid grid-cols-[2fr_1fr_auto_auto] gap-2 items-end">
            <div class="grid gap-2">
              <Label for="newRateName">New Tax Name</Label>
              <Input id="newRateName" v-model="newRate.name" placeholder="e.g. VAT" />
            </div>
            <div class="grid gap-2">
              <Label for="newRateValue">Rate (%)</Label>
              <Input id="newRateValue" v-model.number="newRate.rate" type="number" min="0" step="0.01" />
            </div>
            <div class="flex items-center gap-2 pb-2">
              <Checkbox v-model="newRate.compound" />
              <Label>Compound</Label>
            </div>
            <Button :disabled="isAddingRate" @click="addRate">
              <Plus class="h-4 w-4" /> Add
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </FinanceLayout>
</template>
