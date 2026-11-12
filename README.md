# Monitoramento de Qualidade da Água em Rios (ODS 6 & 14)

Prova de conceito para cadastrar pontos de coleta em rios, lançar medições e ler o alerta quando o valor sai do limiar da estação. A API é em Go, os dados ficam no MongoDB e a interface é em Vue.

Atividade de Estudo Programada (AEP) de Engenharia de Software.

## Alinhamento aos ODS

- **ODS 6 - Água potável e saneamento:** parâmetros como pH, turbidez e oxigênio dissolvido.
- **ODS 14 - Vida na água:** o alerta registra a medição que ficou fora do limiar.

## O que usa

- Go 1.23+
- MongoDB 7.0
- Vue 3, Vite, TypeScript, Vue Router e Pinia
- Swagger / OpenAPI (Swaggo)
- Docker e Docker Compose
- GitHub Actions, só para build, teste e cobertura da API

## Como executar

### Pré-requisitos

- [Docker e Docker Compose](https://www.docker.com/)
- Go 1.23+, para a API fora do container e para os testes. O gerador do Swagger: `go install github.com/swaggo/swag/cmd/swag@v1.16.4`, com `swag` no PATH.
- Node.js 20+, para a interface em `web/`.

### Variáveis de ambiente

```bash
cp .env.example .env
```

No PowerShell: `Copy-Item .env.example .env`. As credenciais do exemplo são só para a máquina local.

### API e banco

Opção A, MongoDB e API em containers:

```bash
make docker-all
# ou: docker compose up -d --build --wait
```

Opção B, só o banco no Docker e a API no terminal:

```bash
make docker-up
make run
```

Sem Make: `docker compose up -d --wait mongodb` e `go run ./cmd/api`.

O MongoDB sobe o replica set local `rs0`. Sem isso, gravar uma coleta e o alerta dela na mesma transação falha. O motivo e o desenho das coleções estão em [docs/banco.md](docs/banco.md).

A API fica em [http://localhost:8080](http://localhost:8080). A documentação interativa fica em [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

### Interface

Com a API no ar, na pasta `web/`:

```bash
npm install
npm run dev
```

A interface abre em [http://localhost:5173](http://localhost:5173). O detalhe de ambiente, CORS e organização do Vue está em [docs/frontend.md](docs/frontend.md).

## Documentação

- [Banco](docs/banco.md): coleções, documentos e por que cada uma é aninhada ou referenciada.
- [Backend](docs/backend.md): testes, cobertura e a pipeline.
- [Frontend](docs/frontend.md): como a interface fala com a API e onde estão os padrões de tela.
