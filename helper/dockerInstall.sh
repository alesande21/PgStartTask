#!/bin/sh

# Обновите список пакетов
sudo apt update

# Установите необходимые пакеты, позволяющие apt использовать репозиторий через HTTPS:
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common

# Добавьте GPG ключ официального репозитория Docker:
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

# Добавьте репозиторий Docker к списку источников пакетов APT:
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Обновите список пакетов с информацией о Docker из нового репозитория:
sudo apt update

# Убедитесь, что установлен пакет Docker из официального репозитория, а не из репозитория Ubuntu:
apt-cache policy docker-ce

# Установите Docker:
sudo apt install -y docker-ce
