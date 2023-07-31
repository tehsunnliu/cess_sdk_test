package logger

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	Log *log.Logger
)

func init() {

	loc, _ := time.LoadLocation("Asia/Kolkata")

	var logpath = "./" + time.Now().In(loc).String() + ".log"

	flag.Parse()
	var file, err1 = os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logpath)
}
