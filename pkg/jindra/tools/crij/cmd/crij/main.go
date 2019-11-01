package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/kesselborn/jindra/pkg/jindra/tools/crij"
)

func callScript(json string, waitOnFail bool, stdoutFile, stderrFile string, args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("error executing %s: %s\n", cmd.String(), err)
	}

	_, err = io.WriteString(stdin, json)
	if err != nil {
		log.Fatalf("error executing %s: %s\n", json, err)
	}
	stdin.Close()

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		log.Printf("error executing script: %s ... execute with options '-wait-on-fail' to leave the container running for 5 more minutes) \n", err)
		if waitOnFail {
			time.Sleep(5 * time.Minute)
		}
		log.Fatalf("error executing %s: %s\n", cmd.String(), err)
	}

	err = ioutil.WriteFile(stdoutFile, outbuf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error writing stdout file %s: %s\n", stdoutFile, err)
	}

	err = ioutil.WriteFile(stderrFile, errbuf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error writing stderr file %s: %s\n", stderrFile, err)
	}

	log.Printf("successfully called %s!\n", cmd.String())

}

func main() {
	prefix := flag.String("env-prefix", "", "only env vars with this prefix will be used -- prefix is separated by a '.' (i.e. prefix for env var 'git.source.url' would be git)")
	waitOnFail := flag.Bool("wait-on-fail", false, "leave container live for 5 more minutes if the script fails (for debugging purposes)")
	semaphoreFile := flag.String("semaphore-file", "", "file to watch ... program will start once this file DOES NOT EXIST")
	envFile := flag.String("env-file", "", "file with simple env variables (no interpolation, no multiline values)")
	ignoreMissingEnvFile := flag.Bool("ignore-missing-env-file", false, "don't file, if provided env file does not exist")
	stdoutFile := flag.String("stdout-file", "/dev/stdout", "where to print resources stdout output")
	stderrFile := flag.String("stderr-file", "/dev/stderr", "where to print resources stdout output")
	debug := flag.Bool("debug", false, "print debugging info")
	flag.Parse()

	if *semaphoreFile == "" {
		fmt.Println("crij -- concourse resource in jindra")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *envFile != "" {
		content, err := ioutil.ReadFile(*envFile)
		if err != nil {
			if os.IsNotExist(err) && !*ignoreMissingEnvFile {
				fmt.Fprintf(os.Stderr, "env file %s does not exist\n", envFile)
				os.Exit(1)
			}
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "error reading env file %s: %s\n", *envFile, err)
				os.Exit(1)
			}
		}
		crij.SimpleEnvFileToEnv(string(content))
	}

	fmt.Printf("waiting for %s to go away, writing stdout to %s, stderr to %s ", *semaphoreFile, *stdoutFile, *stderrFile)
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

	callScript(s, *waitOnFail, *stdoutFile, *stderrFile, flag.Args())

	if *debug {
		fmt.Fprintf(os.Stderr, s)
	}

}
