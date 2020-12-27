# TODO

- Versioning
  - We'll need to do something like download the new ruleset in the same location as the old, hopefully
    taking advantage of the fact that we're using go-getter and it will do smart things by not redownloading
    the entire ruleset. From there we should be able to check the version difference and then just recompile
    all rules.
- Write documentation for all of this
  - How to create a new ruleset(this should also be a command to generate the skeleton)
  - The anatomy of a ruleset folder
  - Why the config file exists and how does it work
  - How to manipulate the config file(turn off entire rulesets and rules)
  - Describe that we check files one by one, you will not catch project errors only single file errors
  - Because file has to be read into memory we might have to skip larger files
- Implement suggestions for rules
  - Rules should provide you with suggestion text and suggestion code.
- Language server
- Enable the ability to turn off entire rulesets and leave the underlying rules settings entact
- Do a full review package by package, PoC means a lot of this code was written hastily
- Write code to skip files larger or equal to 2GB
- When you update a ruleset, it should keep whether that ruleset was on or off and the same with any non
  new rules for that ruleset(aka don't change user settings on update)
- Add nocolor option
- Add concurrency to linting, we should be able to run many rules at the same time for a single file. Allow the user to set this
- Remove the init command and add the check for the base init directories at the start of each command that
  requires them
- Think about allowing a pager view of the humanized output
- Add ability to version lock rulesets
- Fix error reporting everywhere
- Take input from stdin
- Make sure we're outputting to correct fds
- Allow users to add comments to hcl files such that they can suppress some rules/rulesets
- Find fix for when text is too long and our spinner renders poorly.
- Document how to make breaking changes.
- Implement IDS for rules.
