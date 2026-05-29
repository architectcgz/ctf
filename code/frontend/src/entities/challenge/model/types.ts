export const challengeCategories = [
  'web',
  'pwn',
  'reverse',
  'crypto',
  'misc',
  'forensics',
] as const

export const challengeDifficulties = ['beginner', 'easy', 'medium', 'hard', 'insane'] as const

export const challengeStatuses = ['published', 'draft', 'archived'] as const

export const challengeInstanceSharingModes = ['shared', 'per_team', 'per_user'] as const

export type ChallengeCategory = (typeof challengeCategories)[number]
export type ChallengeDifficulty = (typeof challengeDifficulties)[number]
export type ChallengeStatus = (typeof challengeStatuses)[number] | (string & {})
export type ChallengeInstanceSharing = (typeof challengeInstanceSharingModes)[number] | (string & {})

export interface ChallengeDirectoryListItem {
  id: string
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  is_solved: boolean
  tags: string[]
  points: number
  solved_count: number
  total_attempts: number
}

export interface ChallengeMetaSummary {
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  is_solved?: boolean
  attachment_url?: string | null
  tags: string[]
}

export interface ChallengeProfileMetaSummary {
  title: string
  attachment_url?: string | null
  image_id?: string | null
  instance_sharing?: ChallengeInstanceSharing
  created_at?: string
  updated_at?: string
}

export interface ChallengeProfileSummary {
  category?: ChallengeCategory | null
  difficulty?: ChallengeDifficulty | null
  points: number
  status?: ChallengeStatus
}
