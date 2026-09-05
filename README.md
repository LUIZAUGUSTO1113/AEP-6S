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
- [Go 1.22+](https://go.dev/) instalado _(opcional se rodar direto pelo Docker)_

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

Para executar a suíte de testes e gerar a evidência de cobertura exigida ($\ge 70\%$):

```bash
make test
```

Para gerar o relatório visual interativo em HTML:

```bash
make test-html
# Abra o arquivo coverage.html no navegador
```
