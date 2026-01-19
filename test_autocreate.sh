#!/bin/bash

echo "Тестирование системы автосоздания файлов..."

# Создаем временный каталог для теста
TEST_DIR="/tmp/test_agent"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Копируем бинарный файл агента
cp /workspace/PR/agent/agent ./

# Запускаем агент в фоне
echo "Запуск агента для тестирования..."
timeout 5s ./agent &
AGENT_PID=$!

# Ждем немного и проверяем созданные файлы
sleep 2

# Проверяем, были ли созданы файлы
echo "Проверка созданных файлов:"
ls -la

if [ -f "configs/agent-config.yaml" ]; then
    echo "✓ Конфигурационный файл агента создан"
else
    echo "✗ Конфигурационный файл агента НЕ создан"
fi

if [ -f "configs/monitor-config.json" ]; then
    echo "✓ Конфигурационный файл монитора создан"
else
    echo "✗ Конфигурационный файл монитора НЕ создан"
fi

if [ -d "logs" ]; then
    echo "✓ Директория для логов создана"
else
    echo "✗ Директория для логов НЕ создана"
fi

# Останавливаем агент
kill $AGENT_PID 2>/dev/null

# Переходим к тестированию сервера
echo ""
echo "Тестирование сервера..."

SERVER_TEST_DIR="/tmp/test_server"
mkdir -p "$SERVER_TEST_DIR"
cd "$SERVER_TEST_DIR"

# Копируем бинарный файл сервера
cp /workspace/PR/server/server ./

# Запускаем сервер в фоне
echo "Запуск сервера для тестирования..."
timeout 5s ./server &
SERVER_PID=$!

# Ждем немного и проверяем созданные файлы
sleep 2

# Проверяем, были ли созданы файлы
echo "Проверка созданных файлов сервера:"
ls -la

if [ -f "configs/config.yaml" ]; then
    echo "✓ Основной конфигурационный файл сервера создан"
else
    echo "✗ Основной конфигурационный файл сервера НЕ создан"
fi

if [ -f "configs/proxy-config.json" ]; then
    echo "✓ Конфигурационный файл прокси создан"
else
    echo "✗ Конфигурационный файл прокси НЕ создан"
fi

if [ -d "logs" ] && [ -d "assets/bins/monitor" ] && [ -d "assets/bins/proxy" ] && [ -d "dist" ]; then
    echo "✓ Все необходимые директории созданы"
else
    echo "✗ Не все необходимые директории созданы"
fi

# Останавливаем сервер
kill $SERVER_PID 2>/dev/null

echo ""
echo "Тестирование завершено!"