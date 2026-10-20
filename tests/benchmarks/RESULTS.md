# Benchmark Fase 1 — lote Pure-Go (sem CGO)

Data: 2026-10-13 · Máquina: Windows + Docker Desktop (engine Linux) · `.exe` cross-compilado (`GOOS=windows`, `CGO_ENABLED=0`)
Fotos: sintéticas geradas por `go run ./tests/gen` (JPEG q90, 800×600 a 2000×1500, retrato+paisagem) · Logo: `tests/testdata/logo.png` (800×400, PNG com alfa)
Opções: `--pos bottom-right --size 20 --margin 3 --opacity 0.7 --quality 92 --format jpeg` · Workers: NumCPU do host

| Lote | Tempo interno | img/s | Falhas | Observação |
|------|---------------|-------|--------|------------|
| 10   | 105 ms        | ~95   | 0      | inclui startup do `.exe` no wall |
| 100  | 913 ms        | ~110  | 0      | |
| 500  | 4,026 s (wall 4,12 s) | ~124 | 0 | teto do RF-05 |

Integridade:
- `scan` na pasta de saída: 500 encontradas, 0 inválidas (todas as saídas decodificam).
- Pasta de entrada intacta: 500 arquivos, nenhuma escrita (saída recusada se igual à entrada — `EnsureOutDir`).
- Sem temporários vazados (`WriteAtomic` + Rename; 0 arquivos `.tmp-*`).
- Segunda execução sem `--overwrite` gerou `foto_001_1.jpg` (política renomear, RF-06).

Ressalvas honestas (SDD §8):
- Fotos sintéticas comprimem mais rápido que fotos reais de câmera; throughput real deve ser menor — repetir com lote fotográfico na Fase 4.
- I/O via bind-mount do OneDrive é mais lento que SSD local; números acima são throughput relativo, não teto absoluto.
- Pico de memória (RSS) não instrumentado nesta fase; arquitetura (1 arquivo/worker, sem carregar o lote) + medição com `docker stats`/pprof ficam para a Fase 4.
- Testes: `go test ./...` verde (geometria 9 posições, opacidade, scanner, saída atômica, isolamento de falha/pânico por arquivo, cancelamento).

Reproduzir:

```powershell
.\scripts\go.ps1 run ./tests/gen -n 500
.\scripts\build-win.ps1
.\bin\watermark.exe batch --in .\tests\testdata\in --logo .\tests\testdata\logo.png --out .\tests\testdata\out --overwrite
```
