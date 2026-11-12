# Frontend

A interface fica em `web/`. É Vue 3 com Vite, TypeScript, Vue Router e Pinia. Três rotas: Estações, Coletas e Alertas. Não há kit de componentes.

A base da API é `VITE_API_URL`. O padrão é `http://localhost:8080`. Para outra URL, copie `web/.env.example` para `web/.env` e altere essa variável. A API só libera o navegador em `http://localhost:5173`.

## Subir

Com a API no ar, na pasta `web/`:

```bash
npm install
npm run dev
```

A interface abre em [http://localhost:5173](http://localhost:5173). `npm run typecheck` roda `vue-tsc`. O build de produção é `npm run build` e gera `web/dist`. Esse build não entra no Compose nem na pipeline. O Compose sobe MongoDB e a API. O Vite fica no host.

## Onde está o contrato da tela

- [padroes-frontend-vue.md](padroes-frontend-vue.md): stack, tokens, pastas, componentes e acessibilidade.
- [praticas-ui-ux.md](praticas-ui-ux.md): o que cada tela pede, erro, lista, mapa e exclusão.
- [plano-mobile.md](plano-mobile.md): o que muda abaixo de `960px`.
