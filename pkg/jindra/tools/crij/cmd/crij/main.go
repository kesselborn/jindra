package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/kesselborn/jindra/pkg/jindra/tools/crij"
)

func formattedJson(jsonString string, prefix string) string {
	var jsonStruct interface{}
	if json.Unmarshal([]byte(jsonString), &jsonStruct) != nil {
		return jsonString
	}

	res, err := json.MarshalIndent(jsonStruct, prefix, "  ")
	if err != nil {
		return jsonString
	}

	return string(res)
}

func dumpDebugInfo(debugOut, prefix, jsonString string) {
	if debugOut != "" {
		fmt.Fprintf(os.Stderr, "error ... dumping debug information to: %s\n", debugOut)
		debugOutContent := `
#!/bin/sh
# call this script to create an env file and instructions on
# how to call the resource to debug potential errors you saw.
# The environment will be saved in /tmp/env -- adjust it to your
# needs if necessary.
#
# input json for resource was:
# {{.CommentedJSON}}
#
cat<<EOF>{{.EnvFile}}
{{range $_, $envVar := .Env}}
{{$envVar}}
{{end}}
EOF

cat<<EOF

call the resoure:
	{{.Bin}} -env-prefix={{.EnvPrefix}} -semaphore-file=/tmp/does-not-exist -env-file={{.EnvFile}} /opt/resource/in {{.OutDir}}

just print the json that the resource will get via stdin:
	 {{.Bin}} -env-prefix={{.EnvPrefix}} -just-print-json -semaphore-file=/tmp/does-not-exist -env-file={{.EnvFile}} /opt/resource/in {{.OutDir}}

EOF
`
		t := template.Must(template.New("script").Parse(debugOutContent))
		f, err := os.Create(debugOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file for dumping debug information: %s", err)
		}
		defer f.Close()
		t.Execute(f, struct {
			Bin           string
			CommentedJSON string
			EnvPrefix     string
			EnvFile       string
			OutDir        string
			Env           []string
		}{os.Args[0], formattedJson(jsonString, "#"), prefix, "/tmp/env", "/tmp/jindra-resource", os.Environ()})
	}
}

func callScript(jsonString, prefix string, waitOnFail bool, stdoutFile, stderrFile string, debugOut string, args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error executing %s: %s\n", cmd.String(), err)
		os.Exit(1)
	}

	_, err = io.WriteString(stdin, jsonString)
	if err != nil {
		dumpDebugInfo(debugOut, prefix, jsonString)
		fmt.Fprintf(os.Stderr, "error writing json to stdin %s\n", err)
		os.Exit(1)
	}
	stdin.Close()

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	dumpInfo := func() {
		err = ioutil.WriteFile(stdoutFile, outbuf.Bytes(), 0644)
		if err != nil {
			dumpDebugInfo(debugOut, prefix, jsonString)
			fmt.Fprintf(os.Stderr, "error writing stdout file %s: %s\n", stdoutFile, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%s", string(outbuf.Bytes()))

		err = ioutil.WriteFile(stderrFile, errbuf.Bytes(), 0644)
		if err != nil {
			dumpDebugInfo(debugOut, prefix, jsonString)
			fmt.Fprintf(os.Stderr, "error writing stderr file %s: %s\n", stderrFile, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s", string(errbuf.Bytes()))
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("error executing script: %s ... execute with options '-wait-on-fail' to leave the container running for 5 more minutes) \n", err)
		dumpInfo()
		if waitOnFail {
			dumpDebugInfo(debugOut, prefix, jsonString)
			time.Sleep(5 * time.Minute)
		}
		fmt.Fprintf(os.Stderr, "error executing %s: %s\n", cmd.String(), err)
		os.Exit(1)
	}
	dumpInfo()

	fmt.Printf("successfully called %s!\n", cmd.String())

}

func main() {
	prefix := flag.String("env-prefix", "", "only env vars with this prefix will be used -- prefix is separated by a '.' (i.e. prefix for env var 'git.source.url' would be git)")
	waitOnFail := flag.Bool("wait-on-fail", false, "leave container live for 5 more minutes if the script fails (for debugging purposes)")
	semaphoreFile := flag.String("semaphore-file", "", "file to watch ... program will start once this file DOES NOT EXIST")
	envFile := flag.String("env-file", "", "file with simple env variables (no interpolation, no multiline values)")
	ignoreMissingEnvFile := flag.Bool("ignore-missing-env-file", false, "don't file, if provided env file does not exist")
	stdoutFile := flag.String("stdout-file", "/dev/stdout", "where to print resources stdout output")
	stderrFile := flag.String("stderr-file", "/dev/stderr", "where to print resources stdout output")
	debugOut := flag.String("debug-out", "", "dump debugging information into the specified file (*NOTE*: this can contain sensitive data like passwords, etc.)")
	justJSON := flag.Bool("just-print-json", false, "don't execute resource, just print the json that would be passed to the resource")
	deleteEnvFileAfterRead := flag.Bool("delete-env-file-after-read", false, "delete env file after it was read: this can be necessary if the env file resides in the resource directory as resources sometimes demand an empty directory")
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
		if *deleteEnvFileAfterRead {
			if err := os.Remove(*envFile); err != nil {
				fmt.Fprintf(os.Stderr, "error deleting env file: %s -- continuing anyways\n")
			}
		}
		crij.SimpleEnvFileToEnv(string(content))
	}

	fmt.Fprintf(os.Stderr, "waiting for %s to go away, writing stdout to %s, stderr to %s ", *semaphoreFile, *stdoutFile, *stderrFile)
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

	if *justJSON {
		fmt.Println(formattedJson(s, ""))
		os.Exit(0)
	}

	callScript(s, *prefix, *waitOnFail, *stdoutFile, *stderrFile, *debugOut, flag.Args())
}
