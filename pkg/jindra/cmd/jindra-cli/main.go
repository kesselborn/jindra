package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	jindraApi "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	"github.com/kesselborn/jindra/pkg/jindra"
)

func configMap(p jindraApi.JindraPipeline, buildNo int) {
	cm, err := jindra.PipelineRunConfigMap(p, buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to config map for pipeline run: %s", err)
	}

	fmt.Println(interface2yaml(cm))
}

func main() {
	config := flag.String("c", "", "jindra pipeline config")
	help := flag.Bool("h", false, "show help text")
	buildNo := flag.Int("build-no", 42, "build number")

	flag.Usage = func() {
		fmt.Printf(`
Usage: %s [OPTIONS] <command>

Commands:
  cm      : print out configmap

Options: 
`, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	usage := func(success bool) {
		exitCode := 0
		if !success {
			exitCode = 1
			flag.CommandLine.SetOutput(os.Stderr)
		}
		flag.Usage()
		os.Exit(exitCode)
	}

	if config == nil || *help {
		usage(*help)
	}

	yamlData, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("error reading file %s: %s", *config, err)
	}

	p, err := jindra.NewJindraPipeline(yamlData)
	if err != nil {
		log.Fatalf("cannot convert yaml to jindra pipeline: %s", err)
	}

	switch flag.Arg(0) {
	case "cm":
		configMap(p, *buildNo)
	case "":
		fmt.Println("no sub command given")
		usage(false)
	default:
		fmt.Printf("unknown sub command %s\n", flag.Arg(0))
		usage(false)
	}

}

func interface2yaml(x interface{}) string {
	jsonTxt, err := json.Marshal(x)
	if err != nil {
		panic("error marshalling to json: " + err.Error())
	}

	yamlTxt, err := yaml.JSONToYAML(jsonTxt)
	if err != nil {
		panic("error running JSONToYAML: " + err.Error())
	}

	return string(yamlTxt)
}
