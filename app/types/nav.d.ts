export interface NavLink {
  title: string
  link: string
  label?: string
  badge?: string
  icon?: string
}

export interface NavSectionTitle {
  heading: string
}

export interface NavGroup {
  title: string
  icon?: string
  label?: string
  children: NavLink[]
}

export interface NavMenu {
  title: string
  icon?: string
  link: string
}

export declare type NavMenuItems = (NavLink | NavGroup | NavSectionTitle)[]
