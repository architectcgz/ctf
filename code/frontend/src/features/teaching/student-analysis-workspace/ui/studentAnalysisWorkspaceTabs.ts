import type { StudentAnalysisWorkspaceTab } from '../model/useStudentAnalysisPage'

export interface StudentAnalysisWorkspaceTabItem {
  key: StudentAnalysisWorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

export const studentAnalysisWorkspaceTabs: StudentAnalysisWorkspaceTabItem[] = [
  {
    key: 'overview',
    label: '学员画像',
    buttonId: 'student-tab-overview',
    panelId: 'student-overview',
  },
  {
    key: 'recommendations',
    label: '推荐任务',
    buttonId: 'student-tab-recommendations',
    panelId: 'student-recommendations',
  },
  {
    key: 'writeups',
    label: '发布的题解',
    buttonId: 'student-tab-writeups',
    panelId: 'student-writeups',
  },
  {
    key: 'evidence',
    label: '证据链',
    buttonId: 'student-tab-evidence',
    panelId: 'student-evidence',
  },
  {
    key: 'training-records',
    label: '训练记录',
    buttonId: 'student-tab-training-records',
    panelId: 'student-training-records',
  },
]

export function findStudentAnalysisWorkspaceTab(
  activeTab: StudentAnalysisWorkspaceTab
): StudentAnalysisWorkspaceTabItem {
  return (
    studentAnalysisWorkspaceTabs.find((tab) => tab.key === activeTab) ??
    studentAnalysisWorkspaceTabs[0]
  )
}
