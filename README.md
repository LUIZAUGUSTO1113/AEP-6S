# Monitoramento de Qualidade da Água em Rios (ODS 6 & 14)

> **Atividade de Estudo Programada (AEP) — Engenharia de Software**  
> Prova de Conceito (PoC) interdisciplinar para monitoramento contínuo da qualidade da água em rios utilizando Go e Banco de Dados NoSQL (MongoDB).

---

## Alinhamento aos Objetivos de Desenvolvimento Sustentável (ONU)

- **ODS 6 — Água Potável e Saneamento:** Monitoramento de parâmetros críticos (pH, Turbidez, Oxigênio Dissolvido) para preservação de mananciais.
- **ODS 14 — Vida na Água:** Prevenção e controle da poluição hídrica para proteção dos ecossistemas aquáticos.

---

## Tecnologias Utilizadas

- **Linguagem:** Go (Golang 1.23+)
- **Banco de Dados:** MongoDB 7.0 (NoSQL)
- **Documentação Interativa:** Swagger / OpenAPI (Swaggo)
- **Conteinerização:** Docker & Docker Compose
- **Integração Contínua (CI):** GitHub Actions

---

## Como Executar o Projeto

### Pré-requisitos

- [Docker e Docker Compose](https://www.docker.com/) instalados
- [Go 1.23+](https://go.dev/) instalado _(opcional se rodar direto pelo Docker)_

### 1. Configurar variáveis de ambiente

Copie o arquivo de exemplo para criar o `.env`:

```bash
cp .env.example .env
```

### 2. Executar a Aplicação

#### Opção A: Execução Completa via Docker (Recomendado para Avaliação)

Sobe o MongoDB e a API Go em containers conectados:

```bash
make docker-all
# ou: docker compose up --build
```

#### Opção B: Desenvolvimento Local Rápido

Sobe apenas o banco no Docker e executa a API no terminal:

```bash
make docker-up   # Sobe o MongoDB
make run         # Regera a documentação e inicia a API
```

---

## Documentação da API (Swagger)

Com a aplicação rodando, acesse a documentação interativa no navegador:
**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

---

## Testes Automatizados e Cobertura

Os testes usam `testing` e Testify. Service e Controller usam um repositório
simulado; os testes de integração do Repository usam MongoDB real.
A API não precisa estar rodando para executar a suíte.

### Preparar o MongoDB exclusivo de testes

Com Docker Desktop iniciado, execute na raiz do projeto:

```powershell
docker compose -f docker-compose.test.yml up -d --wait
$env:MONGO_TEST_URI = "mongodb://localhost:27018"
```

No Bash, use `export MONGO_TEST_URI="mongodb://localhost:27018"`.
A variável deve ser definida em cada novo terminal. A suíte não carrega o `.env`
da aplicação nem utiliza `MONGO_INITDB_DATABASE`. O container usa a porta 27018
e armazenamento temporário, sem compartilhar o volume do banco da aplicação.
Cada teste de integração cria um banco `aep_test_<id>` e remove somente esse banco
ao finalizar. Uma execução interrompida pode deixar bancos nesse container;
eles desaparecem quando o container for removido.

### Executar a suíte completa e gerar cobertura

```powershell
go test -v "-coverprofile=coverage.out" ./...
go tool cover "-func=coverage.out"
go tool cover "-html=coverage.out" -o coverage.html
```

As aspas preservam os argumentos no PowerShell; no Bash, os comandos obrigatórios
também funcionam sem aspas (`go test -v -coverprofile=coverage.out ./...`,
`go tool cover -func=coverage.out` e `go tool cover -html=coverage.out -o coverage.html`).
Confira se o primeiro comando terminou com sucesso antes de usar os relatórios.
O valor relevante é a linha `total:` de `go tool cover -func`, considerando todos
os pacotes do comando `./...`, sem filtrar arquivos da cobertura.

Atalhos equivalentes:

```bash
make test
```

Para gerar o relatório visual interativo em HTML:

```bash
make test-html
# Abra o arquivo coverage.html no navegador
```

Para abrir no Windows: `Start-Process .\coverage.html`.
Para executar apenas os testes que não exigem MongoDB: `make test-unit`
ou `go test -short -v ./...`. Esse modo exibe os testes de integração como
`SKIP` e **não é a execução usada como evidência da cobertura mínima**.
Na suíte completa, `MONGO_TEST_URI` ausente ou MongoDB indisponível causa falha.

Para remover somente o container de testes:

```powershell
docker compose -f docker-compose.test.yml down
```

### O que é validado

- Service: obrigatoriedade de rio/parâmetro, pH de 0 a 14, datas, operações válidas e propagação de erros.
- Controller: respostas JSON, códigos HTTP, dados inválidos e falhas simuladas do banco.
- Repository: Create, FindAll, FindByID, FindByRiver, Update e Delete no MongoDB; ID duplicado, documentos inválidos e cancelamento.
- Database: montagem da URI, configuração isolada e conexão com o banco de testes.
- Inicialização: `cmd/api/main_test.go` inicia o servidor no processo de teste desse
  pacote, em porta dinâmica, e verifica as rotas reais e o Swagger. O servidor
  termina quando esse processo de teste encerra; não usa a porta 8080 nem o banco da aplicação.

Na verificação local de 07/09/2026, a suíte completa passou com **97,5% de cobertura
total de instruções no Go 1.23.0**, versão definida no projeto, e 98,0% no Go 1.27.1
(Windows, MongoDB de testes). Service, Controller e Repository atingiram 100% em
ambas as versões. Esses valores são evidência dessas execuções; gere novamente
os relatórios após alterações. Cobertura de instruções não significa que todos os
comportamentos possíveis foram testados.

### GitHub Actions

A pipeline usa a versão de Go definida em `go.mod`, inicia um MongoDB descartável
com healthcheck e executa a suíte completa. Credenciais de produção e secrets não
são necessários para esse banco de testes. A execução falha se a cobertura total
for menor que 70%. Os arquivos `coverage.out` e `coverage.html` ficam disponíveis
no artefato `coverage-report` da execução. Os atalhos locais mostram a cobertura;
a verificação automática do limite ocorre no CI.

O roteiro para a gravação está em [docs/roteiro-video.md](docs/roteiro-video.md).
