package config

type Config struct {
	Server   Server
	Database Database
	JWT      JWT
}

type Server struct {
	PORT string
}

type Database struct {
	HOST 	string
	NAME 	string
	PORT	 string
	USER 	string
	PASS string
	TZ 	string
}

type JWT struct {
	SECRET string
	EXP    int
}