# TODO

- Versioning
  - We should allow users to lock their version of particular rulesets. We should be able to do this
  - with a symbol before the version, or a simple attribute that says "locked".
- Rule Remediation
  - Allow remediation to have more than one line
- Write documentation for all of this
  - How do we roll out new versions of this? (include the go-plugin portion and the other software version)
- Language server (gives this the ability to embed this into an IDE free of charge).
- Add nocolor option
- Add concurrency to linting, we should be able to run many rules at the same time for a single file. Allow the user to set this.
- Think about allowing a pager view of the humanized output
- Take input from stdin
- Allow users to add comments to hcl files such that they can suppress some rules/rulesets.
- Can we check terminal size before hand and avoid running the spinner for insufficently small terminals?
  (This causes the spinner to render poorly)
- Generator for new rules, rulesets and other things a user might have to do by hand.
- Formatter's printerror should take an error and expand it into a string, so that we can pass around errors not strings.
- Add timeouts for potentially long actions like downloads
