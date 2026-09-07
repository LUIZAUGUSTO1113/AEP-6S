# Roteiro de demonstração — Rafael

Duração planejada: 2 minutos e 30 segundos. Material para incluir no PDF do grupo.

## Antes de gravar

1. Iniciar o MongoDB da aplicação e a API; abrir http://localhost:8080/swagger/index.html.
2. Iniciar o banco de testes: `docker compose -f docker-compose.test.yml up -d --wait`.
3. No PowerShell, definir `$env:MONGO_TEST_URI = "mongodb://localhost:27018"`.
4. Executar `go test -v "-coverprofile=coverage.out" ./...`, depois `go tool cover "-func=coverage.out"` e `go tool cover "-html=coverage.out" -o coverage.html`. As aspas evitam a divisão dos argumentos no PowerShell.
5. Deixar o terminal na linha `total:`, abrir `coverage.html` e a execução do GitHub Actions correspondente ao commit entregue.
6. A execução verde do GitHub só pode ser apresentada depois que as alterações forem enviadas e a pipeline realmente passar. Não confundir o resultado local com o CI.

## 0:00–0:30 — Problema e ODS

**Tela:** título do projeto e README.

**Fala:** “Nosso projeto é uma prova de conceito para registrar medições da qualidade da água em rios, como pH e turbidez. Ele se relaciona ao ODS 6, pela preservação da água, e ao ODS 14, pela proteção dos ecossistemas aquáticos. Nesta versão, a API armazena e consulta medições. A faixa de pH validada é de zero a quatorze; isso não classifica a água como própria para consumo.”

## 0:30–1:00 — Arquitetura

**Tela:** pastas e arquivos de `internal/sample`.

**Fala:** “A aplicação foi desenvolvida em Go e dividida em camadas. O Controller recebe as requisições HTTP e devolve JSON. O Service valida os dados e coordena as operações. O Repository persiste as amostras em documentos no MongoDB. O Service depende de uma interface, permitindo simular o banco nos testes unitários.”

## 1:00–2:00 — Demonstração no Swagger

**POST /samples:** executar com:

```json
{
  "river": "Rio Demonstração AEP",
  "parameter": "pH",
  "value": 7.2,
  "collected_at": "2026-09-07T12:00:00Z"
}
```

**Fala:** “Vou cadastrar uma medição de pH. A API retorna 201 e gera um identificador.”

Copiar o `id` retornado. Executar **GET /samples** e localizar a amostra.

**Fala:** “A listagem recupera a medição armazenada no MongoDB.”

Em **PUT /samples/{id}**, colar o mesmo ID e enviar o JSON anterior com `value: 8.1`, mantendo a data explícita. Conferir resposta 200, valor alterado e mesmo ID.

**Fala:** “Atualizo a medição preservando seu identificador.”

Executar **DELETE /samples/{id}** com o mesmo ID e conferir 204. Se houver tempo, listar novamente.

**Fala:** “Por fim, excluo a amostra. O retorno 204 indica sucesso sem corpo de resposta.”

## 2:00–2:30 — Testes, cobertura e CI

**Tela:** resultado da suíte, linha `total:`, HTML e GitHub Actions.

**Fala:** “Os testes unitários verificam as regras do Service e os retornos HTTP, incluindo pH inválido e falhas do banco. Os testes de integração verificam as operações do Repository em MongoDB real e isolado. Este relatório mostra a cobertura total desta execução, acima da meta de setenta por cento. A pipeline executa a suíte e reprova a entrega se a cobertura ficar abaixo do limite.”

Só usar a frase “acima da meta” se o relatório exibido confirmar o resultado.
Mostrar a pipeline verde apenas quando a execução correspondente estiver concluída com sucesso.

Referência da execução local de 07/09/2026: 97,5% de cobertura total com Go 1.23.0
(98,0% com Go 1.27.1); Service, Controller e Repository com 100%. O resultado do GitHub Actions ainda precisa
ser confirmado após envio do commit. Atualize a evidência antes da gravação.

## Evidências para o PDF

- Descrição da estratégia: testes unitários com repositório simulado e integração com MongoDB real.
- Comandos de preparação e de cobertura do README.
- Captura do terminal com testes aprovados e percentual total legível.
- Captura do relatório HTML.
- Link e captura da execução verde do GitHub Actions, identificando o commit.
- Este roteiro e o link do vídeo após sua gravação/publicação.

Referências: https://go.dev/doc/tutorial/add-a-test, https://github.com/stretchr/testify e https://go.dev/blog/cover.
