package main

import (
	"log"
	"monitor-agent/internal/config"
	"monitor-agent/internal/ws"
)

func main() {
	// Загрузка конфигурации агента
	cfg, err := config.LoadAgentConfig("configs/agent-config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации агента:", err)
	}

	// Подключение к серверу через WebSocket
	client, err := ws.NewClient(cfg)
	if err != nil {
		log.Fatal("Ошибка подключения к серверу:", err)
	}

	// Запуск клиента
	client.Start()

	// Блокировка основного потока
	select {}
}