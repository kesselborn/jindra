### Annotations
## Pipeline
| Annotation                  | Meaning                                                      |
| --------------------------- | ------------------------------------------------------------ |
| `jindra.io/build-no-offset` | Offset for your build number (if you re-create a pipeline which already had runs before) -- **must be a string!** |



## Stage (Pod)

| Annotation                  | Meaning                                                      |
| --------------------------- | ------------------------------------------------------------ |
| `jindra.io/inputs`          | Comma separated list of input resources of this stage        |
| `jindra.io/outputs`         | Comma separated list of output resources of this stage       |
| `jindra.io/debug-container` | Set to `enable` to have a container with all shared volumes to inspect a stage -- note that this stage can't finish successfully and will never finish when not deleted |
| `jindra.io/services`        | Comma separated list of containers, which provide services for the current stage (e.g. a database for testing) and shouldn't be waited for to finish |
| `jindra.io/outputs-envs`    | a textual addition / modification for output resources. Use it like this:<br>`jindra.io/outputs-envs: |`<br>	`     registry-image.params.image=./image.tar`<br>	`registry-image.source.tag=latest`<br>	`git.source.uri=git@github.com/jindra/jindra` |
|                             |                                                              |


## Notes to self

### Operator commands

    operator-sdk add api        --api-version=jindra.io/v1alpha1 --kind=JindraPipeline
    operator-sdk add controller --api-version=jindra.io/v1alpha1 --kind=JindraPipeline

### Build and deploy operator


## Available resources


See: https://github.com/concourse/concourse/wiki/Resource-Types
