function escapeCSVValue(value: unknown): string {
  const normalized = String(value ?? '')
  if (!/[",\n\r]/.test(normalized)) {
    return normalized
  }
  return `"${normalized.replaceAll('"', '""')}"`
}

function downloadBlobFile(filename: string, blob: Blob): void {
  const objectURL = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = objectURL
  link.download = filename
  link.click()
  window.URL.revokeObjectURL(objectURL)
}

export function downloadCSVFile(
  filename: string,
  rows: Array<Record<string, string | number | boolean>>
): void {
  if (rows.length === 0) {
    return
  }

  const headers = Object.keys(rows[0])
  const lines = [
    headers.map((header) => escapeCSVValue(header)).join(','),
    ...rows.map((row) => headers.map((header) => escapeCSVValue(row[header])).join(',')),
  ]

  downloadBlobFile(
    filename,
    new Blob([`\uFEFF${lines.join('\n')}`], { type: 'text/csv;charset=utf-8;' })
  )
}

export function downloadJSONFile(filename: string, value: unknown): void {
  downloadBlobFile(
    filename,
    new Blob([`${JSON.stringify(value, null, 2)}\n`], { type: 'application/json;charset=utf-8;' })
  )
}
