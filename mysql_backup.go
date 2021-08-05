package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func backup_mysql() {
	if _, err := os.Stat(BACKUP_PATH); os.IsNotExist(err) {
		println("Path to Backup Does not Exist. Creating Path\n")
		err := os.MkdirAll(BACKUP_PATH, os.ModePerm)
		if err != nil {
			println("Error Creating Path to Backup")
			fmt.Printf("%s\n", err)
			return
		}
	}


	db, err := sql.Open("mysql", MYSQL_USER+":"+MYSQL_PASSWORD+"@tcp(127.0.0.1:3306)/"+MYSQL_DATABASE)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	if err != nil {
		log.Fatal(err)
	}

	res, err := db.Query("SHOW DATABASES")

	if err != nil {
		log.Fatal(err)
	}

	var database string

	for res.Next() {
		err := res.Scan(&database)
		if err != nil {
			panic(err)
		}
		if (database == "information_schema" ) || ( database == "performance_schema" ) || ( database == "mysql" ) || ( database == "sys" ) {
			println("Skipping database: "+database+"\n")
		} else {
			println("Backing up database: "+database)
			currentTime := time.Now()

			date := currentTime.Format("2006-01-02")

			filename := "/backups/mysql/"+date+"-"+database+".gz"

			str := "mysqldump -u "+MYSQL_USER+" -p"+MYSQL_PASSWORD+" --databases "+database+" | gzip -c > "+filename

			_, err := exec.Command("bash", "-c", str).Output()
			if err != nil {
				panic(err)
			}

			_, err = exec.Command("bash", "-c", "gpg --no-tty --yes --batch --passphrase "+GPG_PASSPHRASE+" -c "+filename).Output()
			if err != nil {
				panic(err)
			}

			_, err = exec.Command("bash", "-c", "rm "+filename).Output()
			if err != nil {
				panic(err)
			}

			var cutoff = 730 * time.Hour

			fileInfo, err := ioutil.ReadDir(BACKUP_PATH)
			if err != nil {
				log.Fatal(err.Error())
			}
			now := time.Now()
			for _, info := range fileInfo {
				if diff := now.Sub(info.ModTime()); diff > cutoff {
					fmt.Printf("Deleting %s which is %s old\n", info.Name(), diff)
				}
			}
		}
	}

	files, err := WalkMatch(BACKUP_PATH, "*.gz")

	for _, file := range files {
		_, err = exec.Command("bash", "-c", "gpg --no-tty --yes --batch --passphrase "+GPG_PASSPHRASE+" -c "+file).Output()
		if err != nil {
			panic(err)
		} else {
			println("Encrypted "+file)
		}

		_, err = exec.Command("bash", "-c", "rm "+file).Output()
		if err != nil {
			panic(err)
		}
	}

}
