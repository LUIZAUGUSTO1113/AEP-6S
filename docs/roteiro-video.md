# Roteiro da segunda entrega

Duração prevista de 4min10s. Os três trechos devem usar a mesma versão da PoC.

## Antes de gravar

1. Subir a API com o volume novo do Compose e abrir o Swagger em `http://localhost:8080/swagger/index.html`.
2. Subir o Vue com a API em `http://localhost:8080` e a interface em `http://localhost:5173`.
3. Subir o replica set de testes com `docker compose -f docker-compose.test.yml up -d --wait`.
4. No PowerShell, definir `$env:MONGO_TEST_URI = "mongodb://127.0.0.1:27018/?replicaSet=rs0&directConnection=true"`.
5. Executar os três comandos de cobertura do README. Separar a linha `total:`, o HTML e a execução aprovada do Actions referente ao commit da tag.
6. Abrir o GitHub Project e a tag `entrega-2` no navegador.

## 0:00–0:40 · Rafael

Apresentar o problema de acompanhar a qualidade da água em rios, a relação com os ODS 6 e 14 e a evolução desde a primeira entrega. A PoC em Go e MongoDB passou a vincular cada coleta a uma estação e exibir alertas quando os limiares são ultrapassados.

## 0:40–1:20 · Matheus

Mostrar as coleções `stations`, `samples` e `alerts`. Explicar por que localização e limiares ficam aninhados na estação e por que coletas e alertas usam referências por `station_id`.

## 1:20–3:00 · Responsável pelo Vue

Criar uma estação no Rio Ivaí com limiar de pH de 6 a 9. Selecionar a estação para registrar pH 7,2, depois registrar pH 4,2. Mostrar as duas coletas salvas e filtrar os alertas pela estação. A tela deve mostrar o alerta da segunda coleta.

Se precisar reproduzir a sequência no Swagger, copie o `id` de `POST /stations`. Use esse ID como `station_id` em cada `POST /samples` com `parameter: "pH"` e o respectivo `value`. Consulte `GET /alerts?station_id=<id>`.

## 3:00–3:40 · Rafael

Mostrar a suíte aprovada, a linha `total:` com pelo menos 70%, o relatório HTML e o Actions verde do commit entregue. Explicar que há testes unitários e testes de integração em MongoDB isolado, incluindo o caso em que falha a gravação do alerta.

## 3:40–4:10 · Rafael

Mostrar o quadro de tarefas e a tag `entrega-2`. Encerrar com o resultado exibido na PoC. Juntar as gravações, revisar a duração final e incluir no material da entrega o link do vídeo publicado no YouTube.
