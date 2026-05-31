import type { Component } from 'vue'

export type IconComp = Component
export type NavQuery = Record<string, string>

export type NavItem = {
  name: string
  path: string
  title: string
  icon: IconComp
  roles?: string[]
  query?: NavQuery
  children?: NavItem[]
  moduleKey?: string
  variant?: 'default' | 'backoffice-child'
}

export type NavGroup = {
  key: string
  title: string
  shortTitle: string
  items: NavItem[]
}
