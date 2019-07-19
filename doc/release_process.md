# Pravega Operator Release Process

## Release Versioning
Pravega Operator follows the [Semantic Versioning](https://semver.org/) model for numbering releases.

## Introduction
This page documents the tagging, branching and release process followed for Pravega Operator.

## Types of Releases

### Minor Release (Bug Fix release)

This is a minor release with backward compatible changes and bug fixes.

1. Create a new branch with last number bumped up from the existing release branch.
   For example, if the existing release branch is 0.3.2, the new branch will be named 0.3.3.
   
   `$ git clone --branch <tag-name> git@github.com:pravega/pravega-operator.git `
   
   `$ git checkout -b <release-branch-name>`
   
2. Cherry pick commits from master/private branches into the release branch.
   Change operator version in Version.go
   
    `$ git cherry-pick --signoff <commit Id>`
    
3. Make sure all unit and end to end tests pass successfully. 

    `$ make test`
    
4. Push changes to the newly created release branch.

    `$ git push origin <release-branch-name>`
    
5. Create a new release candidate tag on this branch. 
   Tag name should correspond to release-branch-name-<release-candidate-version>. 
   For example: `0.3.3-rc1` for the first release candidate.
   
    `$ git tag -a <tag-name> -m "<description>"`
    
    `$ git push origin <tag-name>`
    
   It is possible that a release candidate is problematic and we need to do a new release candidate. In this case, we need to repeat        this tagging step as many times as needed.
    
6. Push docker image for release to docker hub pravega repo:

    `$ make build-image`
    
    `$ docker tag pravega/pravega-operator:latest pravega/pravega-operator:<tag-name>`
    
    `$ docker push pravega/pravega-operator:<tag-name>`
    
7. Once a release candidate is tested and there are no more changes needed, push a final release tag and image (like `0.3.3`) 

8. Release Notes

### Major Release (Feature + bugfixes)

This has non backward compatible changes. 
Here, we bump up the middle or most significant digit from earlier release.
Follow same steps as minor release.

## Reference
https://github.com/pravega/pravega/wiki/How-to-release


