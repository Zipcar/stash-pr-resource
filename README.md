# Stash PR Resource

A resource for interacting with stash PRs, useful as a CHECK resource to trigger pipelines when PRs are opened or modified.

## Source Configuration

### Exclusive to CHECK

* username - Username to authenticate against Stash.
* password - Password to authenticate against Stash.
* stash_url - The stash root URL without protocol (such as stash.company.com).
* days_back - If non-zero, only include PRs that have commits more recent than this value.  If zero, no limit.
* project_name - The project key (not friendly name) of the stash project.
* repo_name - The name of the repository.
* pronly - Only consider PRs instead of all changes to branches.  Accepts boolean only.
* branches - (Optional) Branches to include (source of PR, not destination).  Accepts regex.
* ignore_branches - (Optional) Branches to ignore (source of PR, not destination).  Accepts regex.
* paths - (Optional) the filepaths within the repo to include.

### Exclusive to IN

* private_key - Used for stash authentication.
* repo - The URL of the destination repo (ending in .git)

### Example

```
- name: pr-resource
  type: stash-pr
  source:
    days_back: 0
    password: weakpassword1234
    private_key: |-
      -----BEGIN RSA PRIVATE KEY-----
      MIIEowIBAAKCAQEAtCS10/f7W7lkQaSgD/mVeaSOvSF9ql4hf/zfMwfVGgHWjj+W
      <more text>
      DWiJL+OFeg9kawcUL6hQ8JeXPhlImG6RTUffma9+iGQyyBMCGd1l
      -----END RSA PRIVATE KEY-----
    project_name: my_project
    pronly: true
    repo: ssh://git@stash.company.com/my_project/my_repo.git
    repo_name: my_repo
    stash_url: stash.company.com
    username: joe.user
 ```

## Building and Testing

### Building the Image and Running Tests

Note that the tests are run as part of the Docker build, and the build will fail if any tests fail.

```
docker build -t stash-pr-resource .
```

### Manual Testing in the Docker Container

```
## Running CHECK:
/opt/resource/check

## At stdin prompt enter something like:
{"source":{"days_back":0,"username":"joe.user","pronly":true,"password":"weakpassword1234","project_name":"my_project","repo_name":"my_repo","stash_url":"stash.company.com"}}
```

```
## Running IN:
/opt/resource/in /tmp/project_directory

## At stdin prompt enter something like:
{"version":{"changed_branch":"branch1","ref":"SHA of branch1","the_branches":"branch2::SHA of branch2,branch1::SHA of branch1"},"source":{"repo":"ssh://stash.company.com/my_project/my_repo","private_key":"|-\n-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAtCS10/f7W7lkQaSgD/mVeaSOvSF9ql4hf/zfMwfVGgHWjj+W\n<more text>\nDWiJL+OFeg9kawcUL6hQ8JeXPhlImG6RTUffma9+iGQyyBMCGd1l\n-----END RSA PRIVATE KEY-----"}}
```

## Contributing

Push changes to github and create a PR. Currently you'll need to build and push to your local registry until supported in the public Docker registry.
