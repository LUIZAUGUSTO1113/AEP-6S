# Backend

A API está em Go. As rotas saem de `cmd/api`. Estação, coleta e alerta ficam em `internal/station`, `internal/sample` e `internal/alert`. O modelo das coleções está em [banco.md](banco.md).

Com a API em execução, o Swagger fica em [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html). `make run` regenera essa documentação antes de subir o processo. O gerador é `github.com/swaggo/swag/cmd/swag@v1.16.4`.

A API aceita o navegador do Vite em `http://localhost:5173`. Outra origem não passa no preflight. A base que a interface usa por padrão é `http://localhost:8080`.

## Testes

Os testes usam `testing` e Testify. Service e Controller usam um repositório simulado. Os testes de integração do Repository usam MongoDB real. A API da porta 8080 não precisa estar rodando.

### MongoDB exclusivo de testes

Com o Docker iniciado, na raiz do projeto:

```powershell
docker compose -f docker-compose.test.yml up -d --wait
$env:MONGO_TEST_URI = "mongodb://127.0.0.1:27018/?replicaSet=rs0&directConnection=true"
```

No Bash: `export MONGO_TEST_URI="mongodb://127.0.0.1:27018/?replicaSet=rs0&directConnection=true"`.

A variável vale para aquele terminal. A suíte não lê o `.env` da aplicação nem usa `MONGO_INITDB_DATABASE`. O container escuta na porta 27018, com armazenamento temporário, separado do volume da aplicação. Cada teste de integração cria um banco `aep_test_<id>` e remove só esse banco ao terminar. Uma execução interrompida pode deixar bancos nesse container. Eles somem quando o container é removido.

Para remover só esse container:

```powershell
docker compose -f docker-compose.test.yml down
```

### Suíte e cobertura

```powershell
go test -v "-coverprofile=coverage.out" ./...
go tool cover "-func=coverage.out"
go tool cover "-html=coverage.out" -o coverage.html
```

As aspas preservam os argumentos no PowerShell. No Bash, os mesmos comandos funcionam sem aspas: `go test -v -coverprofile=coverage.out ./...`, `go tool cover -func=coverage.out` e `go tool cover -html=coverage.out -o coverage.html`.

Confira se o primeiro comando terminou com sucesso antes de ler os relatórios. O valor que importa é a linha `total:` de `go tool cover -func`, com todos os pacotes de `./...`, sem filtrar arquivo.

Atalhos:

```bash
make test
make test-html
```

`make test-html` gera `coverage.html`. No Windows: `Start-Process .\coverage.html`.

`make test-unit`, ou `go test -short -v ./...`, pula o que exige MongoDB. Os testes de integração aparecem como `SKIP`. Essa execução não é a evidência da cobertura mínima. Na suíte completa, `MONGO_TEST_URI` ausente ou MongoDB indisponível faz o teste falhar.

### O que a suíte cobre

- Service: estação existente, parâmetro obrigatório, pH de 0 a 14, datas, operações válidas e propagação de erros.
- Controller: JSON, códigos HTTP, corpo inválido e falha simulada do banco.
- Repository: Create, FindAll, FindByID, FindByRiver, Update e Delete no MongoDB. ID duplicado, documento inválido e cancelamento.
- Estações: documento aninhado, `min ≤ max`, limiar de pH entre 0 e 14, `Evaluate`, exclusão bloqueada quando há coletas (409), código duplicado (409) e CRUD com o índice único.
- Alertas: filtro por estação, um alerta por coleta, reavaliação e exclusão. O teste HTTP de integração verifica o rollback quando o MongoDB rejeita a gravação do alerta, e a disputa entre cadastrar uma coleta e excluir a estação dela.
- CORS: preflight do Vite e recusa de outra origem.
- Database: montagem da URI, configuração isolada e conexão com o banco de testes.
- Inicialização: `cmd/api/main_test.go` sobe o servidor no processo de teste desse pacote, em porta dinâmica, e verifica as rotas reais e o Swagger. O processo encerra o servidor. Não usa a porta 8080 nem o banco da aplicação.

## GitHub Actions

O workflow em `.github/workflows/ci.yml` usa a versão de Go de `go.mod`, sobe o replica set descartável do Compose de testes e roda a suíte completa. Esse banco não precisa de credencial de produção nem de secret. A execução falha se a cobertura total for menor que 70%. `coverage.out` e `coverage.html` ficam no artefato `coverage-report`. Os atalhos locais mostram a cobertura. O limite de 70% é verificado no CI.
