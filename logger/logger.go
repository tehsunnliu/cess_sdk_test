package logger

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

func init() {

	var logpath = "./logs/"
	if _, err := os.Stat(logpath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(logpath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	loc, _ := time.LoadLocation("Asia/Kolkata")
	logpath += time.Now().In(loc).String() + ".log"

	flag.Parse()
	var file, err1 = os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile: " + logpath)
}
