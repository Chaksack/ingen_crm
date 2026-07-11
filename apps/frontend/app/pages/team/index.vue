<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface Member {
  id: string
  email: string
  display_name: string
  role: string
  is_active: boolean
  created_at: string
  business_unit_id: string | null
  business_unit_name: string | null
}
interface Invite {
  id: string
  email: string
  role: string
  expires_at: string
  created_at: string
  expired: boolean
}
interface RoleOption {
  id: string
  name: string
  builtin: boolean
}
interface BusinessUnit {
  id: string
  name: string
  parent_id: string | null
}

const api = useApi()

const members = ref<Member[]>([])
const invites = ref<Invite[]>([])
const roles = ref<RoleOption[]>([])
const businessUnits = ref<BusinessUnit[]>([])
const memberError = ref('')

const inviteOpen = ref(false)
const inviteEmail = ref('')
const inviteRole = ref('Member')
const inviteSubmitting = ref(false)
const inviteError = ref('')
const lastInviteLink = ref('')

async function load() {
  memberError.value = ''
  try {
    const [m, i, r, b] = await Promise.all([
      api<Member[]>('/team/members'),
      api<Invite[]>('/team/invites'),
      api<RoleOption[]>('/team/roles'),
      api<BusinessUnit[]>('/team/business-units'),
    ])
    members.value = m ?? []
    invites.value = i ?? []
    roles.value = r ?? []
    businessUnits.value = b ?? []
  } catch {
    memberError.value = 'Could not load team data.'
  }
}

async function sendInvite() {
  if (!inviteEmail.value.trim()) return
  inviteSubmitting.value = true
  inviteError.value = ''
  try {
    const res = await api<{ token: string }>('/team/invites', {
      method: 'POST',
      body: { email: inviteEmail.value, role: inviteRole.value },
    })
    lastInviteLink.value = `${window.location.origin}/accept-invite?token=${res.token}`
    inviteEmail.value = ''
    inviteRole.value = 'Member'
    await load()
  } catch {
    inviteError.value = 'Could not create invite. That email may already be a member.'
  } finally {
    inviteSubmitting.value = false
  }
}

async function revokeInvite(id: string) {
  await api(`/team/invites/${id}`, { method: 'DELETE' })
  await load()
}

async function setRole(member: Member, role: string) {
  try {
    await api(`/team/members/${member.id}`, { method: 'PATCH', body: { role } })
    await load()
  } catch {
    memberError.value = `Could not change role for ${member.display_name}. The organization must keep at least one active Admin.`
  }
}

async function setBusinessUnit(member: Member, businessUnitId: string) {
  try {
    await api(`/team/members/${member.id}`, {
      method: 'PATCH',
      body: { business_unit_id: businessUnitId === 'none' ? '' : businessUnitId },
    })
    await load()
  } catch {
    memberError.value = `Could not change business unit for ${member.display_name}.`
  }
}

async function setActive(member: Member, isActive: boolean) {
  try {
    await api(`/team/members/${member.id}`, { method: 'PATCH', body: { is_active: isActive } })
    await load()
  } catch {
    memberError.value = `Could not update status for ${member.display_name}. The organization must keep at least one active Admin.`
  }
}

function copyInviteLink() {
  navigator.clipboard.writeText(lastInviteLink.value)
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Team</h1>
        <p class="text-muted-foreground">Manage who has access to this organization.</p>
      </div>
      <Dialog v-model:open="inviteOpen">
        <DialogTrigger as-child>
          <Button @click="lastInviteLink = ''">
            <Icon name="lucide:user-plus" class="mr-1.5 size-4" />
            Invite teammate
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invite a teammate</DialogTitle>
            <DialogDescription>They'll get a link to set their own password and join this organization.</DialogDescription>
          </DialogHeader>
          <form v-if="!lastInviteLink" class="space-y-4" @submit.prevent="sendInvite">
            <div class="space-y-1.5">
              <Label for="invite-email">Email</Label>
              <Input id="invite-email" v-model="inviteEmail" type="email" required />
            </div>
            <div class="space-y-1.5">
              <Label for="invite-role">Role</Label>
              <Select v-model="inviteRole">
                <SelectTrigger id="invite-role"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="Member">Member</SelectItem>
                  <SelectItem value="Admin">Admin</SelectItem>
                </SelectContent>
              </Select>
              <p class="text-xs text-muted-foreground">
                Custom roles can be assigned from the member list after they join.
              </p>
            </div>
            <p v-if="inviteError" class="text-sm text-destructive">{{ inviteError }}</p>
            <Button type="submit" class="w-full" :disabled="inviteSubmitting">
              {{ inviteSubmitting ? 'Creating…' : 'Create invite' }}
            </Button>
          </form>
          <div v-else class="space-y-3">
            <p class="text-sm text-muted-foreground">
              Share this link with the invitee. It expires in 7 days and can only be used once.
            </p>
            <div class="flex gap-2">
              <Input :model-value="lastInviteLink" readonly />
              <Button variant="outline" @click="copyInviteLink">Copy</Button>
            </div>
            <Button variant="ghost" class="w-full" @click="inviteOpen = false">Done</Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>

    <p v-if="memberError" class="text-sm text-destructive">{{ memberError }}</p>

    <Card>
      <CardHeader><CardTitle class="text-base">Members</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Business Unit</TableHead>
              <TableHead>Status</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="m in members" :key="m.id">
              <TableCell>{{ m.display_name }}</TableCell>
              <TableCell class="text-muted-foreground">{{ m.email }}</TableCell>
              <TableCell>
                <Select :model-value="m.role" @update:model-value="(v) => setRole(m, v as string)">
                  <SelectTrigger class="w-40"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="r in roles" :key="r.id" :value="r.name">{{ r.name }}</SelectItem>
                  </SelectContent>
                </Select>
              </TableCell>
              <TableCell>
                <Select :model-value="m.business_unit_id ?? 'none'" @update:model-value="(v) => setBusinessUnit(m, v as string)">
                  <SelectTrigger class="w-44"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">Unassigned</SelectItem>
                    <SelectItem v-for="b in businessUnits" :key="b.id" :value="b.id">{{ b.name }}</SelectItem>
                  </SelectContent>
                </Select>
              </TableCell>
              <TableCell>
                <Badge :variant="m.is_active ? 'secondary' : 'destructive'">
                  {{ m.is_active ? 'Active' : 'Deactivated' }}
                </Badge>
              </TableCell>
              <TableCell class="text-right">
                <Button v-if="m.is_active" variant="ghost" size="sm" @click="setActive(m, false)">Deactivate</Button>
                <Button v-else variant="ghost" size="sm" @click="setActive(m, true)">Reactivate</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Card v-if="invites.length">
      <CardHeader><CardTitle class="text-base">Pending invites</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Expires</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="i in invites" :key="i.id">
              <TableCell>{{ i.email }}</TableCell>
              <TableCell><Badge variant="outline">{{ i.role }}</Badge></TableCell>
              <TableCell class="text-muted-foreground">
                {{ new Date(i.expires_at).toLocaleDateString() }}
                <Badge v-if="i.expired" variant="destructive" class="ml-1">Expired</Badge>
              </TableCell>
              <TableCell class="text-right">
                <Button variant="ghost" size="sm" @click="revokeInvite(i.id)">Revoke</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
