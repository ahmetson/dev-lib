# Dev Lib and Context
The *Dev* module exposes a developer context.

The contexts access to the *config* engine.
And to the *dep* manager.

This is the developer context.
The configuration engine in the developer context is
working with local yaml files.
The dep manager in the developer context is using the local
directory.

> If you run `go test ./...` or `go test ./dep_manager`, then
> then run them with the `-v` flag. Since the test works
> with source code building.

# Dev Context
Which means it's in the current machine.

The dependencies are including the extensions and proxies.

How it works?

The orchestra is set up. It checks the folder. And if they are not existing, it will create them.
>> dev.Run(orchestra)

Then let's work on the extension.
User is passing an extension url.
The service is checking whether it exists in the data or not.
If the service exists, it gets the yaml. 
And return the config.

If the service doesn't exist, it checks whether the service exists in the bin.
If it exists, then it runs it with --build-config.

Then, if the service doesn't exist in the bin, it checks the source.
If the source exists, then it will call `go build`.
Then call bin file with the generated files.

Lastly, if a source doesn't exist, it will download the files from the repository using go-git.
Then we build the binary.
We generate config.

Lastly, the service.Run() will make sure that all binaries exist.
If not, then it will create them.

-----------------------------------------------
The running the application will do the following.
It checks the port of proxies is in use.
If it's not, then it will call a run.

Then it will call itself.

The service will have a command to "shutdown" contexts. As well as "rebuild"