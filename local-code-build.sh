#!/usr/bin/env sh
# source: https://aws.amazon.com/blogs/devops/announcing-local-build-support-for-aws-codebuild/
image="aws/codebuild/golang:1.10"
src="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" # Use the directory of this script as the source directory
artifacts=$src

docker pull amazon/aws-codebuild-local:latest --disable-content-trust=false

docker run -it -v /var/run/docker.sock:/var/run/docker.sock -e "IMAGE_NAME=$image" -e "ARTIFACTS=$artifacts" -e "SOURCE=$src" amazon/aws-codebuild-local