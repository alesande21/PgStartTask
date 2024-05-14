package db

import (
	"PgStartTask/server/api"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type RealDB struct {
	DB *sql.DB
}

func (r *RealDB) Query(query string, args ...interface{}) (api.Commands, error) {
	pErr := r.DB.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return nil, pErr
	}

	var foundCommand api.Command
	var commands []api.Command

	rows, err := r.DB.Query(query, args...)

	if err != nil {
		log.Println("Ошибка при поиске команды!", err)
		return commands, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
		if err != nil {
			log.Println("Ошибка при считывании строк:", err)
			return commands, err
		}
		commands = append(commands, foundCommand)
	}
	return commands, nil
}

func (r *RealDB) Ping() error {
	err := r.DB.Ping()
	if err != nil {
		log.Println("Problems connecting to the database!")
		return err
	}
	return nil
}

func (r *RealDB) QueryRow(query string, args ...interface{}) (*api.Command, error) {
	pErr := r.DB.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return nil, pErr
	}

	var command api.Command
	err := r.DB.QueryRow(query, args...).Scan(&command.Id, &command.BodyScript, &command.ResultRunScript, &command.Status)
	if err != nil {
		//log.Println("Failed to insert row:", err)
		return nil, err
	}

	return &command, nil
}

func (r *RealDB) Exec(query string, args ...interface{}) error {
	pErr := r.DB.Ping()
	if pErr != nil {
		log.Println("Problems connecting to the database!")
		return pErr
	}

	_, err := r.DB.Exec(query, args...)

	return err
}

func ConnectionToDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Println("Problems connecting to the database!")
		panic(err)
	}

	log.Println(api.ColorString(api.FgYellow, "Успешное подключение к базе данных!"))
	//fmt.Println("Successfully connected!")

	return db
}
