import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ContestSummary {
  id: string
  title: string
  status: string
}

export interface ScoreboardRow {
  rank: number
  team: string
  solved: number
  score: number
}

export interface ContestAnnouncement {
  id: string
  title: string
  content?: string
  created_at?: string
}

export const useContestStore = defineStore('contest', () => {
  const currentContest = ref<ContestSummary | null>(null)
  const scoreboard = ref<ScoreboardRow[]>([])
  const announcements = ref<ContestAnnouncement[]>([])
  const isFrozen = ref(false)
  const myTeam = ref<unknown>(null)

  function updateScoreboard(rows: ScoreboardRow[]): void {
    scoreboard.value = rows
  }

  function addAnnouncement(item: ContestAnnouncement): void {
    announcements.value = [item, ...announcements.value]
  }

  function setFrozen(val: boolean): void {
    isFrozen.value = val
  }

  return {
    currentContest,
    scoreboard,
    announcements,
    isFrozen,
    myTeam,
    updateScoreboard,
    addAnnouncement,
    setFrozen,
  }
})
