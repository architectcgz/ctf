import { describe, expect, it } from 'vitest'

import { useStudentFilters } from './useStudentFilters'

describe('useStudentFilters', () => {
  it('应该输出裁剪后的学生查询参数', () => {
    const filters = useStudentFilters({
      selectedClassName: 'Class A',
      searchQuery: ' Alice ',
      studentNoQuery: ' 2024001 ',
    })

    expect(filters.selectedClassName.value).toBe('Class A')
    expect(filters.studentQueryParams.value).toEqual({
      keyword: 'Alice',
      student_no: '2024001',
    })

    filters.updateSearchQuery('   ')

    expect(filters.studentQueryParams.value).toEqual({
      keyword: undefined,
      student_no: '2024001',
    })
  })

  it('应该只重置文本筛选，不清空当前班级', () => {
    const filters = useStudentFilters({
      selectedClassName: 'Class B',
      searchQuery: 'alice',
      studentNoQuery: '2024002',
    })

    filters.resetTextFilters()

    expect(filters.selectedClassName.value).toBe('Class B')
    expect(filters.searchQuery.value).toBe('')
    expect(filters.studentNoQuery.value).toBe('')
    expect(filters.studentQueryParams.value).toEqual({
      keyword: undefined,
      student_no: undefined,
    })
  })
})
