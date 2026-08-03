import { nanoid } from 'nanoid/non-secure'

export interface Project {
  id: string
  key: string
  name: string
  description?: string
  leadId?: string
  memberIds: string[]
  createdAt: number
  customerName?: string
  boards?: string[]
}

export interface Member {
  id: string
  name: string
  avatar?: string
}

const STORAGE_KEY = 'projects.v1'

const defaultMembers: Member[] = [
  { id: 'u_self', name: 'Andrew Chakdahah', avatar: '/avatars/avatartion.png' },
  { id: 'u_1', name: 'IBS Support', avatar: '/avatars/03.png' },
  { id: 'u_2', name: 'Accounting', avatar: '/avatars/02.png' },
]

export function useProjects() {
  const projects = useState<Project[]>('projects', () => load())
  const members = useState<Member[]>('project-members', () => defaultMembers)

  function load(): Project[] {
    if (import.meta.client) {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) {
        try { return JSON.parse(raw) as Project[] }
        catch { }
      }
    }
    return []
  }

  function persist() {
    if (import.meta.client)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(projects.value))
  }

  function createProject(input: { name: string; key?: string; description?: string; leadId?: string; memberIds?: string[]; customerName?: string; boards?: string[] }) {
    const id = nanoid(8)
    const key = (input.key || input.name).toUpperCase().replace(/\s+/g, '-').slice(0, 10)
    const p: Project = {
      id,
      key,
      name: input.name.trim(),
      description: input.description?.trim(),
      leadId: input.leadId,
      memberIds: input.memberIds && input.memberIds.length ? Array.from(new Set(input.memberIds)) : [defaultMembers[0].id],
      createdAt: Date.now(),
      customerName: input.customerName?.trim(),
      boards: input.boards && input.boards.length ? Array.from(new Set(input.boards)) : ['Board'],
    }
    projects.value = [p, ...projects.value]
    persist()
    return p
  }

  function removeProject(id: string) {
    projects.value = projects.value.filter(p => p.id !== id)
    persist()
  }

  function getProject(id: string) {
    return projects.value.find(p => p.id === id)
  }

  function updateProject(id: string, patch: Partial<Pick<Project, 'name' | 'key' | 'description' | 'leadId' | 'memberIds' | 'customerName' | 'boards'>>) {
    const idx = projects.value.findIndex(p => p.id === id)
    if (idx === -1)
      return
    const curr = projects.value[idx]
    const next: Project = { ...curr, ...patch }
    // Normalize key if provided
    if (patch.key !== undefined && patch.key !== null) {
      next.key = String(patch.key).toUpperCase().replace(/\s+/g, '-').slice(0, 10)
    }
    // Ensure name trimmed
    if (patch.name !== undefined && patch.name !== null)
      next.name = String(patch.name).trim()
    if (patch.customerName !== undefined && patch.customerName !== null)
      next.customerName = String(patch.customerName).trim()
    if (patch.boards !== undefined && patch.boards !== null) {
      const boards = Array.isArray(patch.boards) ? patch.boards : []
      next.boards = boards.filter(Boolean)
      if (!next.boards.length)
        next.boards = ['Board']
    }
    projects.value.splice(idx, 1, next)
    persist()
  }

  return { projects, members, createProject, removeProject, getProject, updateProject }
}
