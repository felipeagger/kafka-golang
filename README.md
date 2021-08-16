# kafka-golang
PoC using Apache & Confluent Kafka with Golang

# Subir a Aplicacao com Docker:
  Acesse a raiz do repositorio e rode: 
  
```  
  make docker  
```

  Parar a Aplicacao: make dockerdown  

  Para mais detalhes: make help

# Variaveis de Ambiente:
  
```  
TOPIC: events
GROUP: ConsumerGrpDocker
BROKERS: kafka1:1909,kafka2:2909
LOCALSTACK: http://localhost:4576 
```
# Dashboard Kafka 

Kowl: http://0.0.0.0:8088


# Links

https://github.com/confluentinc/confluent-kafka-go

https://github.com/cloudhut/kowl
