package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"jc.org/immotep/cmd"
)

func main() {

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	cmd.Execute()
}
