## Release guide for team members

To build and publish a new version of `secure-setup-terraform`, including publishing binaries for
all supported OSes and architectures, complete the following steps:

- Find the previously released version. You can do this by looking at the git
  tags or by looking at the frontpage of this repo on the right side under the
  "Releases" section.

- Figure out the version number to use. We use "semantic versioning"
  (https://semver.org), which means our version numbers look like
  `MAJOR.MINOR.PATCH`. Quoting semver.org:

        increment the MAJOR version when you make incompatible API changes
        increment the MINOR version when you add functionality in a backward compatible manner
        increment the PATCH version when you make backward compatible bug fixes

  The most important thing is that if we change an API or command-line user
  journey in a way that could break an existing use-case, we must increment the
  major version.

- Update the `action.yml` to change the `RELEASE_VERSION` variable to the new desired release version
  and submit a pull request (this seems odd, but we want to reference the future version as it will
  be built after we tag this commit).

- After the pull request from the previous step is merged, push a signed tag in git, with
  the tag named with your version number, with a
  message saying why you're creating this release. For example:

        $ git tag -s -a v0.2.7 -m 'Release v0.2.7'
        $ git push origin v0.2.7

- A GitHub workflow will be triggered by the tag push and will handle
  everything. You will see the new release created within a few minutes. If not,
  look for failed [Release workflow runs](https://github.com/abcxyz/secure-setup-terraform/actions/workflows/release.yml)
  and look at their logs.
