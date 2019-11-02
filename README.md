### Annotations
## Pipeline
| Annotation                  | Meaning                                                      |
| --------------------------- | ------------------------------------------------------------ |
| `jindra.io/build-no-offset` | Offset for your build number (if you re-create a pipeline which already had runs before) -- **must be a string!** |



## Stage (Pod)

| Annotation                        | Meaning                                                      |
| --------------------------------- | ------------------------------------------------------------ |
| `jindra.io/inputs`                | Comma separated list of input resources of this stage        |
| `jindra.io/outputs`               | Comma separated list of output resources of this stage       |
| `jindra.io/debug-container`       | Set to `enable` to have a container with all shared volumes to inspect a stage. The stage pod will not finish unless you delete the file `/DELETE_ME_TO_STOP_DEBUG_CONTAINER` withing the container `jindra-debug-container` (where the pipeline continues) or delete the pod (where the pipeline files) |
| `jindra.io/services`              | Comma separated list of containers, which provide services for the current stage (e.g. a database for testing) and shouldn't be waited for to finish |
| `jindra.io/outputs-envs`          | a textual addition / modification for output resources. Use it like this:<br>`jindra.io/outputs-envs: |`<br>	`     registry-image.params.image=./image.tar`<br>	`registry-image.source.tag=latest`<br>	`git.source.uri=git@github.com/jindra/jindra`<br><br>**Note**: no interpolation of other environment variables is done, nor are multiline values supportet; don't put quotes around the value |
| `jindra.io/first-init-containers` | Comma separated list of init-container names that should be executed in the specified order _before_ the jindra-injected init containers (input resources, transit resource). All resource mounts will be available in these init containers. |


## Notes to self

### Operator commands

    operator-sdk add api        --api-version=jindra.io/v1alpha1 --kind=JindraPipeline
    operator-sdk add controller --api-version=jindra.io/v1alpha1 --kind=JindraPipeline

### Build and deploy operator


## Available resources

## Resources

...

- `RESOURCE_DIR/.jindra.resource.env`
- how do parameters work (env -> json)
  - places where to set parameters, order & precedence
Resource outputs are saved in resource folder at:

- `.jindra.in-resource.stderr`
- `.jindra.in-resource.stdout`

- `.jindra.out-resource.stderr`
- `.jindra.out-resource.stdout`


## Debugging

Resource outputs are saved in resource folder at:

- `.jindra.in-resource.stderr`
- `.jindra.in-resource.stdout`

- `.jindra.out-resource.stderr`
- `.jindra.out-resource.stdout`

- debug annotation

See: https://github.com/concourse/concourse/wiki/Resource-Types
