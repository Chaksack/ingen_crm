export interface NavLink {
  title: string
  link: string
  icon?: string
}

export interface NavGroup {
  title: string
  icon?: string
  children: NavLink[]
}

export type NavItem = NavLink | NavGroup
