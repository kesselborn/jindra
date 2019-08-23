## magic vars in pipeline.yaml

  - JINDRA_RESOURCE_DIR: directory in which resources will get mounted
  - JINDRA_TRANSIT: directory where to save artifacts that should move to next stage

  Concourse: https://concourse-ci.org/implementing-resource-types.html#resource-metadata

  - $BUILD_ID
    The internal identifier for the build. Right now this is numeric but it may become a guid in the future. Treat it as an absolute reference to the build.
  - $BUILD_NAME
    The build number within the build's job.
  - $BUILD_JOB_NAME
    The name of the build's job.
  - $BUILD_PIPELINE_NAME
    The pipeline that the build's job lives in.
  - $BUILD_TEAM_NAME
    The team that the build belongs to.
  - $ATC_EXTERNAL_URL
    The public URL for your ATC; useful for debugging.


## Available resources

See: https://github.com/concourse/concourse/wiki/Resource-Types

## Validate
- no owner reference in stages
- restartPolicy: nicht gesetzt oder never

## Contraints to document
- all validations from above
- resources _must_ have a shell available (sh is sufficient)
