# Pascom-MarcaFoto (Watermark Studio)

Aplicativo de marca d'água em lote para Windows.

- **Fase 1** (concluída): Núcleo Pure-Go + CLI Windows (cross-compile via Docker)
- **Fase 2** (em desenvolvimento): Interface desktop Wails v2 + React/TypeScript

---

## Fase 1 — CLI (Pure Go, sem CGO)

### Pré-requisitos
- Docker Desktop (engine Linux) — todo o Go roda em container `golang:1.27`

### Uso
```powershell
# testes (no container)
.\scripts\go.ps1 test ./...

# gerar fotos sintéticas + logo
.\scripts\go.ps1 run ./tests/gen -n 20

# compilar o .exe Windows (cross-compile no container)
.\scripts\build-win.ps1

# varrer, pré-visualizar, aplicar, processar lote
.\bin\watermark.exe scan --in .\tests\testdata\in
.\bin\watermark.exe preview --in .\tests\testdata\in\foto_001.jpg --logo .\tests\testdata\logo.png --out .\tests\testdata\out\prev.jpg
.\bin\watermark.exe batch --in .\tests\testdata\in --logo .\tests\testdata\logo.png --out .\tests\testdata\out
```

Opções do lote: `--pos bottom-right --size 20 --margin 3 --opacity 0.7 --quality 92 --format jpeg|png --workers N [--overwrite]`.
Limite: 500 imagens/execução. A saída nunca pode ser a pasta de entrada; conflitos geram `nome_1.jpg` salvo com `--overwrite`.

---

## Fase 2 — Desktop App (Wails v2 + React)

### Pré-requisitos (Windows nativo)
- **Go 1.22+** (para CGO/webview)
- **Node.js 18+**
- **Wails v2 CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **GCC/MinGW** (via `choco install mingw` ou TDM-GCC) — necessário para CGO no Windows

### Desenvolvimento (hot-reload)
```powershell
# Inicia Vite (frontend) + Wails (backend) com hot-reload
.\scripts\dev-wails.ps1
```

### Build de produção
```powershell
# Compila frontend (Vite) + backend (Go) + empacota .exe Windows
.\scripts\build-wails.ps1
```

Gera: `bin/watermark-studio.exe` (ou `bin/Watermark Studio.exe`)

### Interface
1. **Selecionar Fotografias** — escolhe pasta, escaneia recursivamente (magic bytes), mostra contagem
2. **Marca D'Água** — importa logo (PNG com transparência recomendado), preview da logo
3. **Configurações Visuais** — posição (9 pontos), tamanho %, margem %, opacidade, formato (JPEG/PNG), qualidade, workers
4. **Pasta de Saída** — destino dos arquivos processados (originais preservados)
5. **Pré-visualização** — gera preview reduzido (max 800px) com marca aplicada
6. **Processamento** — barra de progresso, arquivo atual, cancelamento controlado
7. **Relatório** — concluídas, falhas, tempo, pasta de saída, lista de erros

---

## Layout do Projeto

```
watermark-studio/
├── main.go                    # Entry point Wails (Fase 2)
├── wails.json                 # Config Wails
├── go.mod / go.sum
├── cmd/watermark/             # CLI Fase 1
│   └── main.go
├── internal/
│   ├── app/                   # Backend Wails (bindings Go↔TS)
│   │   └── app.go
│   ├── watermark/             # Core: geometria, composição, EXIF
│   │   ├── options.go
│   │   ├── geometry.go
│   │   ├── composite.go
│   │   └── orientation.go
│   ├── files/                 # Scanner, saída atômica, conflitos
│   │   ├── scanner.go
│   │   └── output.go
│   └── batch/                 # Worker pool, progresso, cancelamento
│       └── manager.go
├── frontend/                  # React + TypeScript + Vite (Fase 2)
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── styles.css
│       ├── types.ts
│       └── bindings.ts
├── scripts/
│   ├── go.ps1                 # Wrapper Go no Docker (Fase 1)
│   ├── build-win.ps1          # Cross-compile CLI (Fase 1)
│   ├── dev-wails.ps1          # Dev server Wails (Fase 2)
│   └── build-wails.ps1        # Build produção Wails (Fase 2)
├── tests/
│   ├── gen/                   # Gerador fotos sintéticas
│   ├── testdata/              # Imagens de teste
│   └── benchmarks/            # Resultados de performance
└── bin/                       # Executáveis gerados
```

---

## Arquitetura (Fase 2)

| Camada | Tecnologia | Responsabilidade |
|--------|------------|------------------|
| **Frontend** | React 18 + TypeScript + Vite | UI, estado, eventos, preview base64 |
| **Bridge** | Wails v2 (Go↔TS bindings) | Chamadas assíncronas, eventos em tempo real |
| **Backend** | Go (internal/app) | Orquestra: scanner, watermark, batch, filesystem |
| **Core** | Pure Go (imaging, goexif) | Decode, orientação EXIF, resize, overlay, encode |

### Fluxo de Dados
```
UI (React) → Wails Bindings → App (Go) → internal/watermark + files + batch
                                    ↓
                              Eventos: progress, current, complete, error
                                    ↓
                              UI atualiza em tempo real
```

---

## Testes

```powershell
# Testes unitários (core Go) — roda no container Docker
.\scripts\go.ps1 test ./...

# Testes de integração Wails — requer build nativo
wails test
```

---

## Benchmarks

Ver `tests/benchmarks/RESULTS.md` para métricas de:
- Tempo total por lote (10/100/500 imagens)
- Imagens/segundo
- Pico de memória (RSS)
- Taxa de falhas

---

## Licença

Uso interno Pascom.