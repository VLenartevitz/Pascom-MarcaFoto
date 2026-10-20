// Bindings TypeScript para o backend Go (Wails v2)
// Este arquivo é gerado automaticamente pelo Wails, mas criamos manualmente para tipagem

interface Window {
  backend: Backend;
}

interface Backend {
  // Config
  ValidateConfig(cfg: Config): Promise<string | null>;
  
  // Arquivos
  ScanFolder(path: string): Promise<ScanResult>;
  SelectInputFolder(): Promise<string>;
  SelectLogoFile(): Promise<string>;
  SelectOutputFolder(): Promise<string>;
  SetLogo(path: string): Promise<string | null>;
  GetLogoPath(): Promise<string>;
  SetOutputDir(path: string): Promise<string | null>;
  GetOutputDir(): Promise<string>;
  
  // Preview
  GeneratePreview(photoPath: string, cfg: Config, maxWidth: number): Promise<PreviewResult>;
  
  // Batch
  StartBatch(inputDir: string, cfg: Config): Promise<BatchReport>;
  CancelBatch(): Promise<void>;
  IsProcessing(): Promise<boolean>;
  
  // Utils
  OpenFolder(path: string): Promise<string | null>;
  GetVersion(): Promise<string>;
  ReadFileAsBase64(path: string): Promise<string>;
  
  // Eventos
  OnBatchProgress(callback: (progress: BatchProgress) => void): Promise<void>;
  OnBatchCurrent(callback: (current: string) => void): Promise<void>;
  OnBatchComplete(callback: (report: BatchReport) => void): Promise<void>;
  OnBatchError(callback: (error: string) => void): Promise<void>;
  OnDefaults(callback: (config: Config) => void): Promise<void>;
}

// Re-exportar tipos
type Config = import('./types').Config;
type ScanResult = import('./types').ScanResult;
type PreviewResult = import('./types').PreviewResult;
type BatchProgress = import('./types').BatchProgress;
type BatchReport = import('./types').BatchReport;