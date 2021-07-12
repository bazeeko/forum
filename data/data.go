package data

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" //sql
	uuid "github.com/satori/go.uuid"
)

func db() (database *sql.DB) {
	database, err := sql.Open("sqlite3", "./data/sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	_, err = database.Exec(`
	create table if not exists users ( 
		id         INTEGER primary key AUTOINCREMENT,
		uuid       varchar(64) not null unique,
		name       varchar(255),
		email      varchar(255) not null unique,
		password   varchar(255) not null,
		created_at timestamp not null   
	  );
	  
	  create table if not exists sessions (
		id         INTEGER primary key AUTOINCREMENT,
		uuid       varchar(64) not null unique,
		email      varchar(255),
		user_id    integer references users(id),
		created_at timestamp not null   
	  );
	  
	  create table if not exists threads (
		id         INTEGER primary key AUTOINCREMENT,
		uuid       varchar(64) not null unique,
		topic      text,
		user_id    integer references users(id),
		created_at timestamp not null       
	  );
	  
	  create table if not exists categories (
		title      varchar(255),
		thread_id  integer references threads(id)
	  );
	  
	  create table if not exists posts (
		id         INTEGER primary key AUTOINCREMENT,
		uuid       varchar(64) not null unique,
		body       text,
		user_id    integer references users(id),
		thread_id  integer references threads(id),
		created_at timestamp not null
	  );
	  
	  create table if not exists thread_rating (
		user_id    integer references users(id),
		thread_id  integer references threads(id),
		liked      boolean
	  );
	  
	  create table if not exists post_rating (
		user_id    integer references users(id),
		post_id    integer references posts(id),
		liked      boolean
	  );
	  
	  create table if not exists images (
		id         integer primary key AUTOINCREMENT,
		thread_id  integer references threads(id),
		img_name   varchar(255)
	  );
	`)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func CreateUUID() (Uuid string) {
	var e error
	return uuid.Must(uuid.NewV4(), e).String()

}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
// func createUUID() (uuid string) {
//   u := new([16]byte)
//   _, err := rand.Read(u[:])
//   if err != nil {
//     log.Fatalln("Cannot generate UUID", err)
//   }

//   // 0x40 is reserved variant from RFC 4122
//   u[8] = (u[8] | 0x40) & 0x7F
//   // Set the four most significant bits (bits 12 through 15) of the
//   // time_hi_and_version field to the 4-bit version number.
//   u[6] = (u[6] & 0xF) | (0x4 << 4)
//   uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
//   return
// }

// hash plaintext with SHA-1
// func Encrypt(plaintext string) (cryptext string) {
// 	cryptext = fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
// 	return
// }
