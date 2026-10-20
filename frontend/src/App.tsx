import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Config, ScanResult, PreviewResult, BatchProgress, BatchReport, POSITIONS, FORMATS, PositionValue, FormatValue } from './types';

// Tipos para eventos do Wails
interface WailsEvents {
  on: (event: string, callback: (...args: any[]) => void) => void;
  off: (event: string, callback: (...args: any[]) => void) => void;
  emit: (event: string, ...args: any[]) => void;
}

declare global {
  interface Window {
    backend: any;
    events: WailsEvents;
  }
}

const DEFAULT_CONFIG: Config = {
  position: 'bottom-right',
  sizePercent: 20,
  marginPercent: 3,
  opacity: 0.7,
  quality: 92,
  format: 'jpeg',
  workers: navigator.hardwareConcurrency || 4,
  overwrite: false,
};

function App() {
  // Estado principal
  const [step, setStep] = useState<'select' | 'configure' | 'preview' | 'processing' | 'complete'>('select');
  const [config, setConfig] = useState<Config>(DEFAULT_CONFIG);
  const [inputDir, setInputDir] = useState<string>('');
  const [logoPath, setLogoPath] = useState<string>('');
  const [logoPreview, setLogoPreview] = useState<string>('');
  const [outputDir, setOutputDir] = useState<string>('');
  const [scanResult, setScanResult] = useState<ScanResult | null>(null);
  const [previewImg, setPreviewImg] = useState<PreviewResult | null>(null);
  const [previewPhoto, setPreviewPhoto] = useState<string>('');
  const [batchReport, setBatchReport] = useState<BatchReport | null>(null);
  const [batchProgress, setBatchProgress] = useState<BatchProgress | null>(null);
  const [currentFile, setCurrentFile] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [processing, setProcessing] = useState(false);
  const [version, setVersion] = useState('');

  // Refs
  const previewPhotoRef = useRef<HTMLImageElement>(null);
  const logoPreviewRef = useRef<HTMLImageElement>(null);

  // Inicializar eventos do Wails
  useEffect(() => {
    const events = window.events;
    
    // Carregar configuração padrão
    events.on('defaults', (cfg: Config) => {
      setConfig(cfg);
    });

    // Progresso do lote
    events.on('batch:progress', (progress: BatchProgress) => {
      setBatchProgress(progress);
    });

    events.on('batch:current', (current: string) => {
      setCurrentFile(current);
    });

    events.on('batch:complete', (report: BatchReport) => {
      setBatchReport(report);
      setProcessing(false);
      setStep('complete');
    });

    events.on('batch:error', (err: string) => {
      setError(err);
      setProcessing(false);
    });

    // Obter versão
    window.backend.GetVersion().then(setVersion);

    return () => {
      events.off('defaults', () => {});
      events.off('batch:progress', () => {});
      events.off('batch:current', () => {});
      events.off('batch:complete', () => {});
      events.off('batch:error', () => {});
    };
  }, []);

  // Carregar preview da logo quando caminho muda
  useEffect(() => {
    if (logoPath) {
      window.backend.ReadFileAsBase64(logoPath).then(setLogoPreview).catch(console.error);
      window.backend.SetLogo(logoPath).catch(console.error);
    } else {
      setLogoPreview('');
      window.backend.SetLogo('');
    }
  }, [logoPath]);

  // Carregar pasta de saída salva
  useEffect(() => {
    window.backend.GetOutputDir().then(setOutputDir);
    window.backend.GetLogoPath().then(p => {
      if (p) setLogoPath(p);
    });
  }, []);

  // Handlers de seleção de arquivos
  const handleSelectInputFolder = async () => {
    try {
      const dir = await window.backend.SelectInputFolder();
      if (dir) {
        setInputDir(dir);
        setError('');
        const result = await window.backend.ScanFolder(dir);
        setScanResult(result);
        if (result.count > 0 && !previewPhoto) {
          setPreviewPhoto(result.files[0]);
        }
        if (result.count > 0) setStep('configure');
      }
    } catch (e: any) {
      setError(e.message || 'Erro ao selecionar pasta');
    }
  };

  const handleSelectLogo = async () => {
    try {
      const file = await window.backend.SelectLogoFile();
      if (file) {
        setLogoPath(file);
        setError('');
      }
    } catch (e: any) {
      setError(e.message || 'Erro ao selecionar logo');
    }
  };

  const handleSelectOutputFolder = async () => {
    try {
      const dir = await window.backend.SelectOutputFolder();
      if (dir) {
        await window.backend.SetOutputDir(dir);
        setOutputDir(dir);
        setError('');
      }
    } catch (e: any) {
      setError(e.message || 'Erro ao selecionar pasta de saída');
    }
  };

  // Handler de configuração
  const handleConfigChange = (key: keyof Config, value: any) => {
    setConfig(prev => ({ ...prev, [key]: value }));
    setPreviewImg(null); // Limpar preview ao mudar config
  };

  // Gerar preview
  const handleGeneratePreview = async () => {
    if (!previewPhoto || !logoPath) return;
    setError('');
    try {
      const result = await window.backend.GeneratePreview(previewPhoto, config, 800);
      if (result.error) {
        setError(result.error);
      } else {
        setPreviewImg(result);
      }
    } catch (e: any) {
      setError(e.message || 'Erro ao gerar preview');
    }
  };

  // Iniciar processamento em lote
  const handleStartBatch = async () => {
    if (!inputDir || !logoPath || !outputDir) {
      setError('Preencha todos os campos obrigatórios');
      return;
    }
    if (scanResult && scanResult.count === 0) {
      setError('Nenhuma imagem para processar');
      return;
    }
    setError('');
    setProcessing(true);
    setBatchProgress(null);
    setBatchReport(null);
    setCurrentFile('');
    setStep('processing');
    
    try {
      await window.backend.StartBatch(inputDir, config);
    } catch (e: any) {
      setError(e.message || 'Erro ao iniciar processamento');
      setProcessing(false);
      setStep('configure');
    }
  };

  // Cancelar processamento
  const handleCancelBatch = async () => {
    await window.backend.CancelBatch();
    setProcessing(false);
  };

  // Abrir pasta de saída
  const handleOpenOutput = async () => {
    if (batchReport?.outDir) {
      await window.backend.OpenFolder(batchReport.outDir);
    }
  };

  // Resetar para nova operação
  const handleNewBatch = () => {
    setStep('select');
    setInputDir('');
    setScanResult(null);
    setPreviewPhoto('');
    setPreviewImg(null);
    setBatchReport(null);
    setBatchProgress(null);
    setCurrentFile('');
    setError('');
  };

  // Renderização condicional por etapa
  return (
    <div className="app">
      <header className="header">
        <div className="header-left">
          <h1>Watermark Studio</h1>
          <span className="version">{version}</span>
        </div>
        <div className="header-right">
          <span className="badge">Processamento Local</span>
        </div>
      </header>

      {error && (
        <div className="toast error" onClick={() => setError('')}>
          ⚠ {error}
        </div>
      )}

      <main className="main">
        {/* ETAPA 1: Seleção de fotos */}
        {step === 'select' && (
          <section className="step-card">
            <h2>1. Selecionar Fotografias</h2>
            <p className="step-desc">Escolha a pasta contendo as fotos para aplicar a marca d'água</p>
            
            <button 
              className="btn btn-primary btn-large"
              onClick={handleSelectInputFolder}
            >
              📁 Selecionar Pasta de Entrada
            </button>

            {inputDir && (
              <div className="selected-path">
                <span className="path-label">Pasta selecionada:</span>
                <span className="path-value">{inputDir}</span>
              </div>
            )}

            {scanResult && (
              <div className="scan-result">
                <div className="stat">
                  <span className="stat-value">{scanResult.count}</span>
                  <span className="stat-label">imagens encontradas</span>
                </div>
                {scanResult.ignored > 0 && (
                  <div className="stat warning">
                    <span className="stat-value">{scanResult.ignored}</span>
                    <span className="stat-label">ignoradas</span>
                  </div>
                )}
                {scanResult.invalid > 0 && (
                  <div className="stat error">
                    <span className="stat-value">{scanResult.invalid}</span>
                    <span className="stat-label">inválidas</span>
                  </div>
                )}
              </div>
            )}

            {scanResult && scanResult.count > 0 && (
              <button className="btn btn-secondary" onClick={() => setStep('configure')}>
                Continuar →
              </button>
            )}
          </section>
        )}

        {/* ETAPA 2: Configuração */}
        {step === 'configure' && (
          <>
            <section className="step-card">
              <h2>2. Marca D'Água</h2>
              <p className="step-desc">Importe sua logo (PNG com transparência recomendado)</p>
              
              <button 
                className={`btn ${logoPath ? 'btn-secondary' : 'btn-primary'} btn-large`}
                onClick={handleSelectLogo}
              >
                {logoPath ? '🔄 Trocar Logo' : '📎 Selecionar Logo'}
              </button>

              {logoPath && (
                <div className="logo-preview-wrapper">
                  <img 
                    ref={logoPreviewRef}
                    src={`data:image/png;base64,${logoPreview}`} 
                    alt="Logo preview"
                    className="logo-preview"
                  />
                  <span className="logo-path">{logoPath}</span>
                </div>
              )}
            </section>

            <section className="step-card">
              <h2>3. Configurações Visuais</h2>
              <div className="config-grid">
                <div className="config-group">
                  <label>Posição</label>
                  <select 
                    value={config.position}
                    onChange={(e) => handleConfigChange('position', e.target.value as PositionValue)}
                    className="select"
                  >
                    {POSITIONS.map(p => (
                      <option key={p.value} value={p.value}>{p.label}</option>
                    ))}
                  </select>
                </div>

                <div className="config-group">
                  <label>Tamanho da Logo: {config.sizePercent}%</label>
                  <input
                    type="range"
                    min="1"
                    max="100"
                    value={config.sizePercent}
                    onChange={(e) => handleConfigChange('sizePercent', Number(e.target.value))}
                    className="slider"
                  />
                </div>

                <div className="config-group">
                  <label>Margem: {config.marginPercent}%</label>
                  <input
                    type="range"
                    min="0"
                    max="40"
                    value={config.marginPercent}
                    onChange={(e) => handleConfigChange('marginPercent', Number(e.target.value))}
                    className="slider"
                  />
                </div>

                <div className="config-group">
                  <label>Opacidade: {config.opacity}</label>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={config.opacity}
                    onChange={(e) => handleConfigChange('opacity', Number(e.target.value))}
                    className="slider"
                  />
                </div>

                <div className="config-group">
                  <label>Formato de Saída</label>
                  <select
                    value={config.format}
                    onChange={(e) => handleConfigChange('format', e.target.value as FormatValue)}
                    className="select"
                  >
                    {FORMATS.map(f => (
                      <option key={f.value} value={f.value}>{f.label}</option>
                    ))}
                  </select>
                </div>

                {config.format === 'jpeg' && (
                  <div className="config-group">
                    <label>Qualidade JPEG: {config.quality}</label>
                    <input
                      type="range"
                      min="1"
                      max="100"
                      value={config.quality}
                      onChange={(e) => handleConfigChange('quality', Number(e.target.value))}
                      className="slider"
                    />
                  </div>
                )}

                <div className="config-group">
                  <label>Workers Paralelos: {config.workers}</label>
                  <input
                    type="range"
                    min="1"
                    max={navigator.hardwareConcurrency || 16}
                    value={config.workers}
                    onChange={(e) => handleConfigChange('workers', Number(e.target.value))}
                    className="slider"
                  />
                </div>

                <div className="config-group checkbox-group">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={config.overwrite}
                      onChange={(e) => handleConfigChange('overwrite', e.target.checked)}
                    />
                    <span>Sobrescrever arquivos existentes</span>
                  </label>
                </div>
              </div>
            </section>

            <section className="step-card">
              <h2>4. Pasta de Saída</h2>
              <p className="step-desc">Onde salvar as fotos processadas (originais nunca são alterados)</p>
              
              <button 
                className={`btn ${outputDir ? 'btn-secondary' : 'btn-primary'} btn-large`}
                onClick={handleSelectOutputFolder}
              >
                {outputDir ? '🔄 Alterar Pasta' : '📂 Selecionar Pasta de Saída'}
              </button>

              {outputDir && (
                <div className="selected-path">
                  <span className="path-label">Pasta de saída:</span>
                  <span className="path-value">{outputDir}</span>
                </div>
              )}

              {scanResult && scanResult.count > 0 && logoPath && outputDir && (
                <button className="btn btn-primary btn-large" onClick={() => setStep('preview')}>
                  Pré-visualizar →
                </button>
              )}
            </section>
          </>
        )}

        {/* ETAPA 3: Pré-visualização */}
        {step === 'preview' && (
          <section className="step-card preview-step">
            <h2>Pré-visualização</h2>
            <p className="step-desc">Verifique como a marca d'água ficará na foto</p>
            
            <div className="preview-container">
              {previewImg ? (
                <img
                  ref={previewPhotoRef}
                  src={`data:image/${config.format === 'png' ? 'png' : 'jpeg'};base64,${previewImg.base64}`}
                  alt="Preview com marca d'água"
                  className="preview-image"
                />
              ) : (
                <div className="preview-placeholder" onClick={handleGeneratePreview}>
                  <span>🔍 Clique para gerar pré-visualização</span>
                </div>
              )}
            </div>

            {previewImg && (
              <div className="preview-info">
                <span>{previewImg.width} × {previewImg.height} px</span>
                <span>Reduzido para max 800px de largura</span>
              </div>
            )}

            <div className="preview-actions">
              <button className="btn btn-secondary" onClick={() => setStep('configure')}>
                ← Voltar
              </button>
              <button 
                className="btn btn-primary btn-large"
                onClick={handleStartBatch}
                disabled={processing}
              >
                {processing ? 'Processando...' : `Iniciar Processamento (${scanResult?.count || 0} fotos)`}
              </button>
            </div>
          </section>
        )}

        {/* ETAPA 4: Processando */}
        {step === 'processing' && (
          <section className="step-card processing-step">
            <h2>Processando Lote</h2>
            
            <div className="progress-container">
              {batchProgress && (
                <>
                  <div className="progress-bar">
                    <div 
                      className="progress-fill" 
                      style={{ width: `${batchProgress.total > 0 ? (batchProgress.done / batchProgress.total) * 100 : 0}%` }}
                    />
                  </div>
                  <div className="progress-stats">
                    <span>{batchProgress.done} / {batchProgress.total} concluídas</span>
                    {batchProgress.failed > 0 && (
                      <span className="failed">• {batchProgress.failed} falhas</span>
                    )}
                    <span>• {batchProgress.elapsed.toFixed(1)}s</span>
                  </div>
                </>
              )}
              {currentFile && (
                <div className="current-file">
                  Processando: {currentFile}
                </div>
              )}
            </div>

            <button className="btn btn-danger" onClick={handleCancelBatch}>
              Cancelar
            </button>
          </section>
        )}

        {/* ETAPA 5: Concluído */}
        {step === 'complete' && batchReport && (
          <section className="step-card complete-step">
            <h2>✅ Processamento Concluído</h2>
            
            <div className="report-grid">
              <div className="report-stat success">
                <span className="report-value">{batchReport.done}</span>
                <span className="report-label">Concluídas</span>
              </div>
              {batchReport.failed > 0 && (
                <div className="report-stat error">
                  <span className="report-value">{batchReport.failed}</span>
                  <span className="report-label">Falhas</span>
                </div>
              )}
              <div className="report-stat">
                <span className="report-value">{batchReport.total}</span>
                <span className="report-label">Total</span>
              </div>
              <div className="report-stat">
                <span className="report-value">{batchReport.elapsed.toFixed(1)}s</span>
                <span className="report-label">Tempo</span>
              </div>
            </div>

            {batchReport.failures.length > 0 && (
              <details className="failures-list">
                <summary>Ver falhas ({batchReport.failures.length})</summary>
                <ul>
                  {batchReport.failures.map((f, i) => (
                    <li key={i}>
                      <code>{f.input.split('\\').pop() || f.input.split('/').pop()}</code>
                      <span>{f.error}</span>
                    </li>
                  ))}
                </ul>
              </details>
            )}

            <div className="complete-actions">
              <button className="btn btn-primary btn-large" onClick={handleOpenOutput}>
                📂 Abrir Pasta de Saída
              </button>
              <button className="btn btn-secondary" onClick={handleNewBatch}>
                Novo Lote
              </button>
            </div>
          </section>
        )}
      </main>

      <footer className="footer">
        <p>Watermark Studio — Fase 2 | Processamento 100% local, sem envio de dados</p>
      </footer>
    </div>
  );
}

export default App;