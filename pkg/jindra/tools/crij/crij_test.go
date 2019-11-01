package crij

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func setEnv(env map[string]string) {
	os.Clearenv()
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func errMsg(exp, got string, err error) string {
	if err != nil {
		return fmt.Sprintf("unexpected error: %s\n", err)
	}
	return fmt.Sprintf("error: expected \n|%s|\n\ngot:\n|%s|\n", exp, got)
}

func TestSimple(t *testing.T) {
	setEnv(map[string]string{"git.foo.bar": "baz"})
	exp := `{
  "foo": {
    "bar": "baz"
  }
}`

	if res, err := EnvToJSON("git"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestEmbeddedArray(t *testing.T) {
	setEnv(map[string]string{"git.foo.bar": "[\"baz\",\"foo\",\"too\"]"})
	exp := `{
  "foo": {
    "bar": [
      "baz",
      "foo",
      "too"
    ]
  }
}`

	if res, err := EnvToJSON("git"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestEmbeddedMap(t *testing.T) {
	setEnv(map[string]string{"git.foo.bar": "{\"baz\":\"foo\",\"too\":\"bad\"}"})
	exp := `{
  "foo": {
    "bar": {
      "baz": "foo",
      "too": "bad"
    }
  }
}`

	if res, err := EnvToJSON("git"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestMultiple(t *testing.T) {
	setEnv(map[string]string{"docker-image.foo.bar": "baz", "docker-image.foo.baz": "bar", "docker-image.bar.foo": "baz", "docker-image.a.b.c": "1"})
	exp := `{
  "a": {
    "b": {
      "c": 1
    }
  },
  "bar": {
    "foo": "baz"
  },
  "foo": {
    "bar": "baz",
    "baz": "bar"
  }
}`

	if res, err := EnvToJSON("docker-image"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestMultipleWithoutPrefix(t *testing.T) {
	setEnv(map[string]string{"foo.bar": "baz", "foo.baz": "bar", "bar.foo": "baz", "a.b.c": "1"})
	exp := `{
  "a": {
    "b": {
      "c": 1
    }
  },
  "bar": {
    "foo": "baz"
  },
  "foo": {
    "bar": "baz",
    "baz": "bar"
  }
}`

	if res, err := EnvToJSON(""); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestMixedEnv(t *testing.T) {
	setEnv(map[string]string{"git.foo.bar": "bar", "docker.foo.baz": "less"})
	exp := `{
  "foo": {
    "bar": "bar"
  }
}`

	if res, err := EnvToJSON("git"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestVersion(t *testing.T) {
	setEnv(map[string]string{"git.versions": "[ { \"ref\": \"61cbef\" }, { \"ref\": \"d74e01\" }, { \"ref\": \"7154fe\" } ]"})
	exp := `{
  "versions": [
    {
      "ref": "61cbef"
    },
    {
      "ref": "d74e01"
    },
    {
      "ref": "7154fe"
    }
  ]
}`

	if res, err := EnvToJSON("git"); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestRootElements(t *testing.T) {
	setEnv(map[string]string{"root": "boo", "foo.": "bar", "versions.": "[ { \"ref\": \"61cbef\" }, { \"ref\": \"d74e01\" }, { \"ref\": \"7154fe\" } ]"})
	exp := `{
  "foo": "bar",
  "versions": [
    {
      "ref": "61cbef"
    },
    {
      "ref": "d74e01"
    },
    {
      "ref": "7154fe"
    }
  ]
}`

	if res, err := EnvToJSON(""); res != exp || err != nil {
		t.Errorf(errMsg(res, exp, err))
	}
}

func TestEnvFileToEnv(t *testing.T) {
	content := `
		  slack.params.text=Job succeeded
		  slack.params.icon_emoji=":ghost:"
		  slack.params.attachments='[{"color":"#00ff00","text":"hihihi"}]'
		  rsync.params.foo="bar"
		  rsync.source.url=rsync://foo.bar
  `
	SimpleEnvFileToEnv(content)

	for _, test := range []struct {
		got         string
		expectation string
		desc        string
	}{
		{os.Getenv("slack.params.text"), "Job succeeded", "env var should be set correctly"},
		{os.Getenv("slack.params.icon_emoji"), `":ghost:"`, "env var should be set correctly"},
		{os.Getenv("slack.params.attachments"), `'[{"color":"#00ff00","text":"hihihi"}]'`, "env var should be set correctly"},
		{os.Getenv("rsync.params.foo"), `"bar"`, "env var should be set correctly"},
		{os.Getenv("rsync.source.url"), "rsync://foo.bar", "env var should be set correctly"},
	} {
		if !reflect.DeepEqual(test.expectation, test.got) {
			t.Errorf(errMsg(test.expectation, test.got, nil))
		}
	}
}
