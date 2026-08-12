export function parseCSV(text: string): string[][] {
  const rows: string[][] = [];
  const lines = text.trim().split("\n");
  for (const line of lines) {
    const cells: string[] = [];
    let current = "";
    let inQuotes = false;
    for (let i = 0; i < line.length; i++) {
      const char = line[i];
      if (char === '"' && inQuotes && line[i + 1] === '"') {
        current += '"';
        i++;
      } else if (char === '"') {
        inQuotes = !inQuotes;
      } else if (char === "," && !inQuotes) {
        cells.push(current);
        current = "";
      } else {
        current += char;
      }
    }
    cells.push(current);
    rows.push(cells);
  }
  return rows;
}

export function expectCSVHeaders(csvText: string, expectedHeaders: string[]): void {
  const rows = parseCSV(csvText);
  if (rows.length === 0) {
    throw new Error("CSV is empty");
  }
  const headers = rows[0];
  for (const expected of expectedHeaders) {
    if (!headers.includes(expected)) {
      throw new Error(`Expected CSV header "${expected}" not found. Headers: ${headers.join(", ")}`);
    }
  }
}

export function expectCSVRowCount(csvText: string, minRows: number): void {
  const rows = parseCSV(csvText);
  if (rows.length < minRows + 1) {
    throw new Error(`Expected at least ${minRows} data rows, got ${rows.length - 1}`);
  }
}
