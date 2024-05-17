# Загрузите последнюю версию Docker Compose:
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Дайте права на выполнение файла:
sudo chmod +x /usr/local/bin/docker-compose

# Создайте символьную ссылку на файл docker-compose, чтобы можно было вызывать его с командой docker-compose:
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

