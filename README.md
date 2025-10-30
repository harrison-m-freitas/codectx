# codectx — Contexto de código
[![CI](https://github.com/harrison-m-freitas/codectx/actions/workflows/ci.yml/badge.svg)](https://github.com/harrison-m-freitas/codectx/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/badge/go-%3E%3D%201.22-blue)](https://go.dev/dl/)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-lightgrey.svg)](./UNLICENSE)

Ferramenta CLI em **Go ≥ 1.22** para coletar **contexto de código** de um ou mais diretórios, aplicando filtros, ordenações e formatos de saída otimizados para leitura humana (plain/markdown/fenced) e para pipelines (JSON/NDJSON). Foco em **desempenho**, **determinismo** e **segurança por padrão**.

> Gere “pacotes de contexto” para prompts, revisões de PR, auditorias rápidas e automações.


## Principais recursos

- **Git-aware**: dentro de um repositório usa `git ls-files -co --exclude-standard` (respeita `.gitignore`); fora disso, varre o FS.
- **Filtros poderosos**:
  - `--ext` (CSV), `--exclude`/`--include` (substring; `-I/--ignore-case` opcional),
  - **segurança** por padrão (oculta `.env`, chaves, tokens, etc.),
  - **binários** ignorados por NUL,
  - `--max-bytes`, `--max-lines`, `--max-cols`.
- **Ordenação determinística**: `path|ext|size|mtime` + processamento concorrente com preservação de ordem.
- **Formatos de saída**: `plain`, `markdown`, `fenced`, `json` (array) e `ndjson` (linhas).
- **Clipboard**: `-C/--clipboard` copia o arquivo final via `atotto/clipboard` ou ferramentas do SO.
- **Limite de arquivos**: `-N/--max-files` trunca e retorna `exit code 3` (sem erro fatal).
- **Logs estruturados**: níveis `ERROR..TRACE`, modo `json`, cor automática e opcional log em arquivo.

## Instalação

### Build local
```bash
go build ./cmd/codectx       # gera ./codectx
````

### Instalação via `go install`

```bash
go install github.com/harrison-m-freitas/codectx/cmd/codectx@latest
```

## Uso rápido

```bash
# Markdown para arquivo
./codectx -p src -p tests -F markdown -o context.out

# Filtra extensões, ordena por tamanho, trunca linhas/colunas
./codectx -p . -e "go,md" -O size --max-lines 400 --max-cols 200

# Apenas ver o que entraria (sem ler conteúdo)
./codectx -p backend --dry-run

# Saída para stdout, já copiando para o clipboard
./codectx -p . -o - -C -F fenced
```

**JSON / NDJSON** (para pipelines):

```bash
# Índice (sem conteúdo) em array JSON
./codectx -p . -F json --index-only > index.json

# Índice (sem conteúdo) em NDJSON (1 JSON por linha)
./codectx -p . -F ndjson --index-only | jq -c '.path'
```

## Semântica dos filtros (resumo)

1. **Extensão** (`--ext`): mantém somente extensões listadas (CSV, case-insensitive).
2. **Excludes/Includes**: correspondência por **substring** no caminho normalizado (`/`), com `-I` para case-insensitive.
3. **Segredos** (`SecretsStrict` padrão **true**): oculta `.env`, chaves, tokens, etc.
4. **Binários** (`BinarySkip` padrão **true**): oculta arquivos com byte NUL.
5. **Tamanho** (`--max-bytes`): ignora arquivos grandes.
6. **Ordem determinística**: coleta → ordena → processa concorrente → imprime por índice.

## Exemplos de saída

**plain** (trecho):

```
================================================================================
FILE: ./internal/cli/flags.go
SIZE: 8703 bytes | HASH: 6127b6ae | LINES: 285
--------------------------------------------------------------------------------
package cli

import (
    "errors"
    ...
```

**markdown** (trecho):

```markdown
## internal/cli/flags.go

 - **Size:** 8703 bytes
 - **Hash:** 6127b6ae
 - **Lines:** 285

  #```
  package cli
  ...
  #```
```

## Testes

```bash
# Executar a suíte toda
go test ./...

# Rodar um teste específico
go test -run '^TestMaxFilesTruncatesAndSignals$' ./internal/scan
```

> Dica: limpe o cache se trocar nomes/arquivos de teste:
> `go clean -testcache`

## Logs por ambiente

* `LOG_LEVEL=0..4` (`0=ERROR .. 4=TRACE`)
* `LOG_TS=0|1` (timestamp)
* `LOG_COLOR=auto|always|never` (auto respeita `NO_COLOR`)
* `LOG_JSON=0|1` (loga em JSON no stderr)
* `LOG_FILE=/caminho/arquivo.log` (duplica log em arquivo)

## Códigos de saída

* `0`: sucesso
* `1`: erro de execução (I/O, validação, etc.)
* `3`: **truncado** por `--max-files` (processou somente os N primeiros)

## Compatibilidade

* Linux, macOS, Windows/WSL. Caminhos normalizados com `/`.

## Licença

[Unlicense](./UNLICENSE) (domínio público).
