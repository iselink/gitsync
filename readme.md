# GitSync

Git sync is a *simple* (refactoring and further improvements required) program for
synchronization of files via git between multiple devices.  

The primary use of this is for me synchronization of obsidian notes to my cloud.  
You can try it for other things, I guess.

Due to simplicity, there is no release — build it.

## Caveats and not implemented things

Well...
- auth via keys is not implemented
- better control of the process...
- refactoring due to shift in implementation
- error handling (if it can't push, it's not going to do anything)

## Configuration

Daemon uses HCL (hashicorp configuration language) as a file format.  
Format itself is quite simple, blocks `entry` can be repeated multiple times with different folders.  
Block `auth` is optional, there is `http` type only available.

```hcl
sync = "10s"
log_level = "info"

entry {
  local_path = "<Path on local filesystem>"
  repository = "http(s) to the git repository"

  auth {
    type = "http"
    username = "your username (if it matter)"
    password = "your password or most likely PAT (personal access token)"
  }
}
```
