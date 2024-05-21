#вход в докер
sudo docker run --rm -it 1c372b3c022e sh

# вход в образ
sudo docker exec -it slava19v-app-1 sh

# посмотреть логи
sudo docker logs slava19v-app-1

#проверка в каких сетях находятся контейнеры
sudo docker network connect slava19v_default slava19v-app-1
sudo docker network connect slava19v_default slava19v-db-1
