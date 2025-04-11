
# goexpert-rate-limiter

Pós Go Expert - Desafio Técnico de Conclusão - Rate Limiter

## Estrutura do projeto
```
.
├── cmd
│   └── rate-limiter
│   │   └── main.go         # Ponto de início do server
├── internal
│   ├── limiter
│   │   └── limiter.go      # Lógica do Limiter
│   ├── middleware
│   │   └── ratelimiter.go  # Middleware GIN Gonic
│   └── storage
│       ├── redis.go        # Implementação do Storage em Redis
│       └── storage.go      # Definição da Interface para permitir trocar o Redis por outro storage
├── Dockerfile
├── README.md
├── docker-compose.yaml
├── go.mod
└── go.sum
```
No arquivo main.go é instanciada o servidor GIN Gonic e toda a configuração do middleware, nesta ordem:
 - Criado um novo `RedisStorage`
 - Criado o `Limiter` (lógica do rate limiter) onde é passado `RedisStorage`, que por sua vez implementa `StorageInterface`
 - O Middleware `RateLimiter` é injetado no servidor GIN, que por sua vez recebe como parâmetro o Limiter e os parâmetros de tempo e limite
 - Desta forma é possível usar outro mecanismo de persistência, bastando este novo mecanismo implementar `StorageInterface`
## Configuração
No arquivo .env (incluso no repositório) é possível definir estes valores:

- REDIS_ADDR (Endereço do Redis)
- REDIS_PASSWORD (Senha Redis)
- DEFAULT_IP_LIMIT (Limite padrão por IP (req/s))
- DEFAULT_TOKEN_LIMIT (Limite padrão por token (req/s))
- BLOCK_DURATION_SECONDS (Tempo de bloqueio após exceder (s))
- WINDOW_DURATION (Duração da janela de tempo para contagem de requisições (s))
```
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
DEFAULT_IP_LIMIT=10
DEFAULT_TOKEN_LIMIT=12
BLOCK_DURATION_SECONDS=30
WINDOW_DURATION=1
```
> Neste exemplo requisições por IP que passarem de 10 por segundo, ficarão
> bloqueadas por 30 segundos.
## Fluxo do Rate Limiter
- Na primeira requisição, a chave é criada no Redis e o TTL é definido com (`WINDOW_DURATION`).
- Nas chamadas subsequentes, é somente incrementado o valor. Assim o TTL permanece o mesmo desde a primeira chamada.
- Fazendo chamadas dentro do limite, a chave irá expirar e ser criada novamente na próxima requisição reiniciando o ciclo.
- Se em uma requisição o valor da chave passar de `DEFAULT_IP_LIMIT / DEFAULT_TOKEN_LIMIT`, é chamada a função de bloqueio. Esta função cria uma chave específica com TTL de  `BLOCK_DURATION_SECONDS`
- Assim todas as chamadas subsequentes retornarão HTTP 429 até que a chave de bloqueio expire.
## Rodando e Testando na prática
1. Clonar o repositório e subir os serviços
```
git clone https://github.com/flaviojohansson/goexpert-rate-limiter
cd goexpert-rate-limiter
docker compose up -d
```
2. Este shellzinho vai extrapolar o limite de 10 req/s
```
for i in {1..20}; do curl http://localhost:8080/user/${i}; echo; sleep .05; done
```
3. É isso que você deverá ver:
```
{"msg":"Olá 1"}
{"msg":"Olá 2"}
{"msg":"Olá 3"}
{"msg":"Olá 4"}
{"msg":"Olá 5"}
{"msg":"Olá 6"}
{"msg":"Olá 7"}
{"msg":"Olá 8"}
{"msg":"Olá 9"}
{"msg":"Olá 10"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}
```
4. O IP foi bloqueado por 30s. Após este tempo será possível tentar de novo.
5. Limpe toda a bagunça
```
docker compose down -v
```