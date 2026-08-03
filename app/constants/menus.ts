import type { NavMenu, NavMenuItems } from '~/types/nav'

export const navMenu: NavMenu[] = [

  {
    title: 'Dashboard',
    icon: 'i-lucide-home',
    link: '/dashboard',
  },
  {
    title: 'Chat',
    icon: 'i-lucide-message-square',
    link: '/chat',
  },
  {
    title: 'Clients',
    icon: 'i-lucide-users',
    link: '/clients',
  },
  {
    title: 'Staff',
    icon: 'i-lucide-user-cog',
    link: '/staff',
  },
  {
    title: 'Projects',
    icon: 'i-lucide-kanban',
    link: '/projects',
  },
  {
    title: 'Human Resource',
    icon: 'i-lucide-briefcase',
    link: '/hrm',
  },
  {
    title: 'Finance',
    icon: 'i-lucide-landmark',
    link: '/finance',
  },
  {
    title: 'Vendors',
    icon: 'i-lucide-handshake',
    link: '/vendor',
  },
]

export const navMenuBottom: NavMenuItems = [
  {
    title: 'Emails',
    label: '9',
    icon: 'i-lucide-mail',
    link: '/email',
  },
  {
    title: 'Customer Support',
    label: '12',
    icon: 'i-lucide-headset',
    link: '/customer-support',
  },
  {
    title: 'OPs Support',
    label: '19',
    icon: 'i-lucide-circle-help',
    link: '/kanban',
  },
]
