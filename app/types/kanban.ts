export interface Task {
  id: string
  title: string
  description?: string
  priority?: 'low' | 'medium' | 'high'
  // Support multiple assignees similar to Jira
  assignees?: Array<{
    id: string
    name: string
    avatar?: string
  }>
  // Scheduling
  startDate?: Date | number | string
  dueDate?: Date | number | string
  status?: string
  labels?: string[]
  createdAt: Date | number | string
  // Subtasks/checklist
  subtasks?: Array<{ id: string; title: string; done: boolean }>
  // Basic attachments metadata for uploads/images
  attachments?: Array<{ id: string; name: string; size: number; type: string; isImage: boolean; url?: string }>
  // Simple comments thread
  comments?: Array<{
    id: string
    author: { id: string; name: string; avatar?: string }
    body: string
    createdAt: Date | number | string
    attachments?: Array<{ id: string; name: string; size: number; type: string; isImage: boolean; url?: string }>
  }>
  parentId?: string | null
  order?: number
}

export interface NewTask extends Omit<Task, 'id' | 'createdAt'> {
}

export interface Column {
  id: string
  title: string
  tasks: Task[]
}

export interface BoardState {
  columns: Column[]
}
