package jindra

import (
	"fmt"
	"path"
	"reflect"
	"testing"

	jindra "github.com/kesselborn/jindra/pkg/apis/jindra/v1alpha1"
	core "k8s.io/api/core/v1"
)

func TestBasicUnmarshalingTest(t *testing.T) {
	p := getExamplePipeline(t)

	for i, test := range []struct {
		got         interface{}
		expectation interface{}
		desc        string
	}{
		{p.Kind, "JindraPipeline", "correct kind"},
		{p.APIVersion, "jindra.io/v1alpha1", "correct api version"},
		{p.ObjectMeta.Name, "http-fs", "correct name"},
		{len(p.Spec.Resources.Triggers), 1, "number of triggers"},
		{p.Spec.Resources.Triggers[0].Name, "git", "trigger name correct"},
		{len(p.Spec.Resources.Containers), 3, "number of resources"},
		{p.Spec.Resources.Containers[2].Name, "slack", "resource name"},
		{len(p.Spec.Stages), 2, "number of stages"},
		{p.Spec.Stages[1].ObjectMeta.Name, "build-docker-image", "stage name"},
		{p.Spec.OnSuccess.ObjectMeta.Annotations["jindra.io/outputs"], "slack", "on success outputs"},
		{p.Spec.OnError.ObjectMeta.Annotations["jindra.io/outputs"], "slack", "on success outputs"},
	} {
		if reflect.DeepEqual(test.expectation, test.got) {
			t.Logf("\t%2d: %-80s %s", i, test.desc, ok())
		} else {
			t.Fatalf("\t%2d: %-80s %s", i, test.desc, errMsg(t, test.expectation, test.got))
		}
	}

}

func TestJob(t *testing.T) {
	job, _ := PipelineRunJob(getExamplePipeline(t), 42)
	expected := *jobFileContents(path.Join(fixtureDir, "jindra.http-fs.42.yaml"), t)

	if !reflect.DeepEqual(expected, job) {
		t.Fatalf("\t%2d: %-80s %s", 0, "job should be correct", errMsg(t, expected, job))
	}
}

func TestConfigMap(t *testing.T) {
	configMap, _ := PipelineRunConfigMap(getExamplePipeline(t), 42)
	expected := *configMapFileContents(path.Join(fixtureDir, "jindra.http-fs.42.stages.yaml"), t)

	if !reflect.DeepEqual(expected, configMap) {
		t.Fatalf("\t%2d: %-80s %s", 0, "config map should be correct", errMsg(t, expected, configMap))
	}
}

func TestRsyncSSHSecret(t *testing.T) {
	secret, _ := NewRsyncSSHSecret(getExamplePipeline(t), 42)
	expected := *secretFileContents(path.Join(fixtureDir, "jindra.http-fs.42.rsync-keys.yaml"), t)

	// we need to cheat a little bit here, as the keys themselves will always differ
	// but we want to test the structure nevertheless
	expected.Data[rsyncSecretPrivateKey] = secret.Data[rsyncSecretPrivateKey]
	expected.Data[rsyncSecretPubKey] = secret.Data[rsyncSecretPubKey]

	if !reflect.DeepEqual(expected, secret) {
		t.Fatalf("\t%2d: %-80s %s", 0, "rsync ssh secret should be correct", errMsg(t, expected, secret))
	}
}

func TestStageConfigs(t *testing.T) {
	configs, configsErr := generateStagePods(getExamplePipeline(t), 42)

	for i, test := range []struct {
		got         interface{}
		expectation interface{}
		desc        string
	}{
		{configsErr, nil, "configs creation should not error out"},
		{len(configs), 5, "pipeline should have four configs"},
		{configs["01-build-go-binary.yaml"], *podFileContents(path.Join(fixtureDir, "jindra.http-fs.42.01-build-go-binary.yaml"), t), "stage 01 should be correct"},
		{configs["02-build-docker-image.yaml"], *podFileContents(path.Join(fixtureDir, "jindra.http-fs.42.02-build-docker-image.yaml"), t), "stage 02 should be correct"},
		{configs["03-on-success.yaml"], *podFileContents(path.Join(fixtureDir, "jindra.http-fs.42.03-on-success.yaml"), t), "on success should be correct"},
		{configs["04-on-error.yaml"], *podFileContents(path.Join(fixtureDir, "jindra.http-fs.42.04-on-error.yaml"), t), "on error should be correct"},
		{configs["05-final.yaml"], *podFileContents(path.Join(fixtureDir, "jindra.http-fs.42.05-final.yaml"), t), "final should be correct"},
	} {
		if reflect.DeepEqual(test.expectation, test.got) {
			t.Logf("\t%2d: %-80s %s", i, test.desc, ok())
		} else {
			t.Fatalf("\t%2d: %-80s %s", i, test.desc, errMsg(t, test.expectation, test.got))
		}
	}

}

func TestWaitForDebugContainer(t *testing.T) {
	ppl := getExamplePipeline(t)
	for _, stage := range ppl.Spec.Stages {
		if stage.Name == "build-docker-image" {
			stage.ObjectMeta.Annotations[jindra.ServicesAnnotationKey] = ""
		}
	}

	configs, _ := generateStagePods(ppl, 42)
	waitForAnnotation := configs["02-build-docker-image.yaml"].ObjectMeta.Annotations[jindra.WaitForAnnotationKey]

	expected := "build-docker-image,jindra-debug-container"

	if !reflect.DeepEqual(expected, waitForAnnotation) {
		t.Fatalf("\t%2d: %-80s %s", 0, "config map should be correct", errMsg(t, expected, waitForAnnotation))
	}

}

func TestInvalidFirstContainer(t *testing.T) {
	ppl := getExamplePipeline(t)
	ppl.Spec.Stages[0].Annotations[jindra.FirstInitContainers] = "xxxxxxxx"

	_, err := generateStagePods(ppl, 42)
	expected := fmt.Errorf("error constructing init containers: defined firstInitContainer xxxxxxxx not found in pipeline definition")

	if !reflect.DeepEqual(expected, err) {
		t.Fatalf("\t%2d: %-80s %s", 0, "incorrect first init container annotation should yield error", errMsg(t, expected.Error(), err.Error()))
	}

}

func TestAnnotationEnvConverter(t *testing.T) {
	annotation := `
		  slack.params.text=Job succeeded
		  slack.params.icon_emoji=":ghost:"
		  slack.params.attachments='[{"color":"#00ff00","text":"hihihi"}]'
		  rsync.params.foo="bar"
		  rsync.source.url="rsync://foo.bar"
	`

	expectedSlackEnv := []core.EnvVar{
		{Name: "slack.params.text", Value: "Job succeeded"},
		{Name: "slack.params.icon_emoji", Value: "\":ghost:\""},
		{Name: "slack.params.attachments", Value: "'[{\"color\":\"#00ff00\",\"text\":\"hihihi\"}]'"},
	}

	e := annotationToEnv(annotation)

	if !reflect.DeepEqual(expectedSlackEnv, e["slack"]) {
		t.Fatalf("\t%2d: %-80s %s", 0, "converting env annotation failed", errMsg(t, expectedSlackEnv, e))
	}
}

func TestDefaultModifier(t *testing.T) {
	pplExpected := getExamplePipeline(t)
	pplUnmodified := getExamplePipeline(t)

	// remove all values that get modified by the modifying webhook
	pplExpected.Annotations[jindra.BuildNoOffsetAnnotationKey] = "0" // adjust to the default value that will get used
	pplExpected.Spec.Final.Name = "foobar"                           // tests that given names are not overwritten
	pplExpected.Spec.Resources.Triggers[0].Schedule = "/5 * * * *"
	pplExpected.Spec.Stages[0].Spec.RestartPolicy = core.RestartPolicyNever
	pplExpected.Spec.Stages[1].Spec.RestartPolicy = core.RestartPolicyNever

	delete(pplUnmodified.Annotations, jindra.BuildNoOffsetAnnotationKey)
	pplUnmodified.Spec.Resources.Triggers[0].Schedule = ""
	pplUnmodified.Spec.OnSuccess.Name = ""
	pplUnmodified.Spec.OnError.Name = ""
	pplUnmodified.Spec.Final.Name = "foobar"
	pplUnmodified.SetDefaults()

	if !reflect.DeepEqual(pplExpected, pplUnmodified) {
		t.Fatalf("\t%2d: %-80s %s", 0, "incorrect first init container annotation should yield error", errMsg(t, pplExpected, pplUnmodified))
	}

}
