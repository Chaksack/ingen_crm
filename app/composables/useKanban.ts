import type { BoardState, Column, NewTask, Task } from '~/types/kanban'
import { useProjects } from '~/composables/useProjects'

const isClient = import.meta.client

const STORAGE_KEY = 'kanban.board.v1'
const FALLBACK_TASK_PREFIX = 'TASK'
function generateTaskId(projectId: string) {
  const { getProject } = useProjects()
  const proj = getProject(projectId)
  const prefix = proj?.key || FALLBACK_TASK_PREFIX
  const counterKey = `kanban.task.counter:${projectId}`
  let counter = 1
  if (import.meta.client) {
    const saved = localStorage.getItem(counterKey)
    counter = saved ? Number.parseInt(saved) + 1 : 1
    localStorage.setItem(counterKey, counter.toString())
  }
  return `${prefix}-${counter.toString().padStart(3, '0')}`
}

function generateColumnId(title: string) {
  return title.toLowerCase().replace(/\s+/g, '-')
}

const defaultBoard: BoardState = {
  columns: [
    { id: 'todo', title: 'To Do', tasks: [] },
    { id: 'in-progress', title: 'In Progress', tasks: [] },
    { id: 'done', title: 'Done', tasks: [] },
  ],
}

export function useKanban(projectId: string = 'default', boardId: string = 'default') {
  // namespace by projectId and boardId to support multiple boards per project
  const stateKey = `kanban-board-${projectId}-${boardId}`
  const storageKey = `${STORAGE_KEY}:${projectId}:${boardId}`
  const board = useState<BoardState>(stateKey, () => load())

  onMounted(() => {
    board.value = load()
  })

  function load(): BoardState {
    if (isClient) {
      const raw = localStorage.getItem(storageKey)
      if (raw) {
        try { return JSON.parse(raw) as BoardState }
        catch { }
      }
      // Fallback to legacy key without boardId for backward compatibility
      const legacyKey = `${STORAGE_KEY}:${projectId}`
      const legacyRaw = localStorage.getItem(legacyKey)
      if (legacyRaw) {
        try { return JSON.parse(legacyRaw) as BoardState }
        catch { }
      }
    }
    return defaultBoard
  }

  function persist() {
    if (isClient) {
      localStorage.setItem(storageKey, JSON.stringify(board.value))
    }
  }

  function addColumn(title: string) {
    const newCol: Column = { id: generateColumnId(title), title, tasks: [] }
    board.value.columns.push(newCol)
    persist()
  }

  function removeColumn(id: string) {
    board.value.columns = board.value.columns.filter(c => c.id !== id)
    persist()
  }

  function updateColumn(id: string, title: string) {
    const col = board.value.columns.find(c => c.id === id)
    if (!col)
      return
    col.title = title
    persist()
  }

  function addTask(columnId: string, payload: NewTask) {
    const col = board.value.columns.find(c => c.id === columnId)
    if (!col)
      return
    col.tasks.unshift({ id: generateTaskId(projectId), createdAt: new Date(), ...payload })
    persist()
  }

  function updateTask(columnId: string, taskId: string, patch: Partial<Task>) {
    const col = board.value.columns.find(c => c.id === columnId)
    if (!col)
      return
    const t = col.tasks.find(t => t.id === taskId)
    if (!t)
      return
    Object.assign(t, patch)
    persist()
  }

  function removeTask(columnId: string, taskId: string) {
    const col = board.value.columns.find(c => c.id === columnId)
    if (!col)
      return
    col.tasks = col.tasks.filter(t => t.id !== taskId)
    persist()
  }

  function setColumns(next: Column[]) {
    board.value.columns = next
    persist()
  }

  return { board, addColumn, removeColumn, updateColumn, addTask, updateTask, removeTask, setColumns, persist }
}
