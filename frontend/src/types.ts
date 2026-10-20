// Tipos compartilhados entre frontend e backend

export interface Config {
  position: string;
  sizePercent: number;
  marginPercent: number;
  opacity: number;
  quality: number;
  format: 'jpeg' | 'png';
  workers: number;
  overwrite: boolean;
}

export interface ScanResult {
  files: string[];
  count: number;
  ignored: number;
  invalid: number;
  error?: string;
}

export interface PreviewResult {
  base64: string;
  width: number;
  height: number;
  error?: string;
}

export interface BatchProgress {
  total: number;
  done: number;
  failed: number;
  elapsed: number;
  current: string;
}

export interface BatchFailure {
  input: string;
  error: string;
}

export interface BatchReport {
  total: number;
  done: number;
  failed: number;
  elapsed: number;
  outDir: string;
  failures: BatchFailure[];
}

export const POSITIONS = [
  { value: 'top-left', label: 'Canto superior esquerdo' },
  { value: 'top-center', label: 'Topo centralizado' },
  { value: 'top-right', label: 'Canto superior direito' },
  { value: 'middle-left', label: 'Meio esquerdo' },
  { value: 'center', label: 'Centro' },
  { value: 'middle-right', label: 'Meio direito' },
  { value: 'bottom-left', label: 'Canto inferior esquerdo' },
  { value: 'bottom-center', label: 'Base centralizada' },
  { value: 'bottom-right', label: 'Canto inferior direito' },
] as const;

export const FORMATS = [
  { value: 'jpeg', label: 'JPEG (com perdas, menor arquivo)' },
  { value: 'png', label: 'PNG (sem perdas, transparência)' },
] as const;

export type PositionValue = typeof POSITIONS[number]['value'];
export type FormatValue = typeof FORMATS[number]['value'];