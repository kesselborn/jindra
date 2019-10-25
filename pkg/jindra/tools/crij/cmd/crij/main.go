package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/kesselborn/jindra/pkg/jindra/tools/crij"
)

func callScript(json string, waitOnFail bool, args []string) {
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
		log.Printf("error executing script: %s ... execute with options '-wait-on-fail' to leave the container running for 5 more minutes) \noutput was: %s\n", err, out)
		if waitOnFail {
			time.Sleep(5 * time.Minute)
		}
		return
	}

	log.Printf("successfully called %s!\noutput was: %s\n", cmd.String(), out)

}

func main() {
	prefix := flag.String("env-prefix", "", "only env vars with this prefix will be used -- prefix is separated by a '.' (i.e. prefix for env var 'git.source.url' would be git)")
	waitOnFail := flag.Bool("wait-on-fail", false, "leave container live for 5 more minutes if the script fails (for debugging purposes)")
	semaphoreFile := flag.String("semaphore-file", "", "file to watch ... program will start once this file DOES NOT EXIST")
	debug := flag.Bool("debug", false, "print debugging info")
	flag.Parse()

	if *semaphoreFile == "" {
		fmt.Println("crij -- concourse resource in jindra")
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

	s, err := crij.EnvToJSON(*prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error converting env to json: %s\n", err)
		os.Exit(1)
	}

	callScript(s, *waitOnFail, flag.Args())

	if *debug {
		fmt.Fprintf(os.Stderr, s)
	}

}
