# Plan

pretty much a copy of tfvet with simplier structure

plugins dir should be separated by rulesets

- aws-basic-ruleset
  - Each ruleset should have a list of plugins that are included in the ruleset
  - Each plugin should provide some documentation, the check function, and its name
- aws-security-ruleset
- aws-infrastructure-team-ruleset

rulesets should exist as separate directories. a cli command should be able to add them.
Once they are added they get put into the config file(we should just write this in hcl?)

- To figure out versioning we will unfortunately need to download and diff files. There isn't an
  easier way without locking down where plugins can be added from.

* Each provider can just put rules one by one in a .go file

- They can provide the documentation via function and link to web documentation
- We build by downloading the repository on behalf of the user and building it on the machine manually
- We ask the plugin for which version it is
- To register the function that the user writes we can probably just accept a function and register it using the SDK
- We want to be able to have the user do this: tflint ruleset docs <name> and have the documentation for the rule
- When we add a ruleset we should note which hash the precompiled file came with, this allows us to save cycles when
  updating.
- Make the SDK limit people to 50 character names
- User can opt out of rules but by default all rules will be used.
- we only check files with .tf extensions and only look in the current directory
- You can also manually pass a list of file paths to check(useful for cis)

SDK functions:

- statusOk, statusFailed
- newTFlintRule(name string, f func() result)
- result {status, errors[line, error, severity]}

function functionhere() {

}

tf.newTFLintRule("howdafuckdoIdothis", runfunctionhere, docfunctionhere)

We could have the

If I make the boundary on the ruleset and single plugin can I ask you for the documentation still?
If I make the boundary on each individual thing a plugin of its own

1. We should download the repository
2. Look in a predefined location for the rules
3. Make some assumptions about the ruleset name and such based on the name of the repository
4. Version for the overall repository determines whether the app will attempt to download and compile new rulesets
5. compile rules into plugins that can be used, note down what were the names of those plugins(plugins on compile can export their name)
6. save plugins under ruleset folder and name them by what their name was
7. when the system starts up we have a main config file in which the user has predefined which rules they do not want
8. When a user runs tflint we read in the file and allow the parsers to run over it by passing it as an argument with go-plugin/grpc
9. We collate answers and display the results of the lint at the end.

- Need to test for private repos

# CMD

PLUGIN_DIR

## TODO

- Plugin system
- Language Server

* Document how to use the configuration file
* Document how rulesets and rules are added
* Document why the ruleset.hcl file has to exist
* Disclose limitation of single file processing. Will not catch global errors
* Do a full review, this code is shit because it PoC
* You cannot do global checking because plugins are called for each file not a grouping of files
* We should have an create function for creating the base things for a new ruleset

* When we read in the plugins when someone wants to run the linter we have to read in each binary and take the name from that.
* If we do that and we track enables and disables we'll have to at that time do a diff, such that we can
* maintain that the ruleset state on disk is the source of truth.
* Add a nocolor option both as the env var and as a command line option

* We should parrellize the linting a group of files, we can have all threads report back to a single
* structure and then wait till they finish to print.
* Add the check for the conf file at the start of every command(maybe not all of them I dunno)--remove the init
* WE make a rules directory so that we don't mistakenly get caught up in any helper directories, make sure to mention that is why the repo rules directory exists
* We can optionally have the user use a pager for extremely large outputs
* Add the ability to lock versions via some sort of syntax

//TODO(clintjedwards): Fix error reporting everywhere
//TODO(clintjedwards): Do a full testing run
//TODO(clintjedwards): Before passing the file to the plugins make sure it compiles/parses
//TODO(clintjedwards): Only run rules which are enabled in the cfg
//TODO(clintjedwards): We should provide a way to take input from stdin, so you can pass it
// parts of files instead of having to read directly from files
//TODO(clintjedwards): Make sure things are being passed correctly to stdout and stderr
//TODO(clintjedwards): We should make a directive at the top of hcl files in the comments, such that users
// can suppress alerts they don't want to apply to this file.

FAQ:

- Make sure you're using hashicorp v2 packages if you're following the documentation on rulesets
