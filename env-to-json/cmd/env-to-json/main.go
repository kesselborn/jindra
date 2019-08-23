package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	envToJSON "github.com/kesselborn/jindra/env-to-json"
)

func callScript(json string, args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("error executing %s: %s\n", cmd.String(), err)
		return
	}

	_, err = io.WriteString(stdin, json)
	if err != nil {
		log.Printf("error executing %s: %s\n", json, err)
		return
	}
	stdin.Close()

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error executing %s: %s\noutput was: %s\n", json, err, out)
		return
	}

	log.Printf("successfully called %s!\noutput was: %s\n", cmd.String(), out)

}

func main() {
	prefix := flag.String("prefix", "", "env var prefix to include in conversion (must be set)")
	semaphoreFile := flag.String("semaphore-file", "", "file to watch ... program will start once this file DOES NOT EXIST")
	debug := flag.Bool("debug", false, "print debugging info")
	flag.Parse()

	if *semaphoreFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("waiting for %s to go away ", *semaphoreFile)
	for {
		_, err := os.Stat(*semaphoreFile)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			fmt.Fprintf(os.Stderr, "error stating %s: %s, continuing anyways", *semaphoreFile, err)
		}
		fmt.Printf(".")
		time.Sleep(1 * time.Second)
	}
	fmt.Println(" done")

	s, err := envToJSON.EnvToJSON(*prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error converting env to json: %s\n", err)
		os.Exit(1)
	}

	callScript(s, flag.Args())

	if *debug {
		fmt.Fprintf(os.Stderr, s)
	}

}
