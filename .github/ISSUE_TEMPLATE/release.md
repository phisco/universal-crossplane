---
name: Release
about: Cut a UXP release
labels: release
---

<!--
Issue title should be in the following format:

    Cut vX.Y.0-up.1 Release on DATE

For example:

    Cut v1.3.0-up.1 on June 29, 2021.

Please assign the release manager to the issue.
-->

This issue can be closed when we have completed the following steps (in order).
Please ensure all artifacts (PRs, workflow runs, Tweets, etc) are linked from
this issue for posterity. Assuming `vX.Y.0-up.1` is being cut, after upstream
[crossplane/crossplane][upstream-xp] `vX.Y.0` has been released
according to the declared [schedule][uxp-schedule], you should have:

- [ ] Check the `release-vX.Y` release branch in [upbound/crossplane][upbound-xp-fork] has been automatically created and is up-to-date, with the upstream [crossplane/crossplane][upstream-xp] release branch, at least up to the `vX.Y.0` tag, adding any required change specific to the fork, see [here][sync-xp-fork] for more details.
- [ ] Cut [upbound/crossplane][upbound-xp-fork] `vX.Y.0-up.1` release from the `release-X.Y` branch by:
  - [ ] Running the [Tag workflow][tag-xp-fork] on the `release-vX.Y` branch with the proper release version, `vX.Y.0-up.1`. Message suggested but not required: `Release vX.Y.0-up.1`.
  - [ ] Running the [CI workflow][ci-xp-fork] on the `release-vX.Y` branch to build and publish the latest tagged artifacts.
  - [ ] You should now be able to run: `docker pull upbound/crossplane:vX.Y.0-up.1`
- [ ] Created and merged a PR to the `main` branch of [upbound/universal-crossplane][uxp] with the following changes, **taking care to label it as `backport release-X.Y`**:
  - [ ] Update any reference to the old latest release to `vX.Y.0-up.1`, such as `CROSSPLANE_TAG` and `CROSSPLANE_COMMIT` in the `Makefile`.
  - [ ] Run `make helm.prepare` to import any change to the templates in the [upstream Helm chart][upstream-helm-chart].
  - [ ] Manually diff and sync [upstream][upstream-xp-values]'s and [uxp][uxp-values]'s `values.yaml.tmpl` as needed, taking care to change any required templating reference, e.g. `%%CROSSPLANE_TAG%%` instead of `%%VERSION%%`. E.g. `export RELEASE_BRANCH=release-X.Y; vimdiff https://raw.githubusercontent.com/upbound/crossplane/$RELEASE_BRANCH/cluster/charts/crossplane/values.yaml.tmpl cluster/charts/universal-crossplane/values.yaml.tmpl`.
  - [ ] Run `make olm.build` to generate the [OLM] bundle.
- [ ] Created the `release-X.Y` branch from `main` branch in [UXP][uxp].
- [ ] Cut [UXP][uxp] `vX.Y.0-up.1` release from the `release-X.Y` branch by:
  - [ ] Running the [Tag workflow][tag-uxp] on the `release-vX.Y` branch with the proper release version, `vX.Y.0-up.1`. Message suggested but not required: `Release vX.Y.0-up.1`.
  - [ ] Running the [CI workflow][ci-uxp] on the `release-vX.Y` branch to build and publish the latest tagged artifacts.
- [ ] Cut the next prerelease of [UXP][uxp] from the `main` branch, `vX.<Y+1>.0-up.1-rc.1` by:
  - [ ] Running the [Tag workflow][tag-uxp] on the `main` branch with the proper release version, `vX.<Y+1>.0-up.1-rc.1`. Message suggested but not required: `Release vX.<Y+1>.0-up.1-rc.1`.
- [ ] Run the [Promote workflow][promote-uxp] to promote `vX.Y.0-up.1` to [stable][uxp-stable-channel], it should contain `universal-crossplane-X.Y.0-up.1.tgz`. Verify everything is correctly working by running `up uxp install` against an empty Kubernetes cluster, e.g. `kind create cluster`, which should result in an healthy UXP installation with expected image versions.
- [ ] Created and published well authored release notes for [UXP][uxp-releases] `vX.Y.0-up.1`. See the previous release for an example, these should at least:
  - [ ] enumerate relevant updates that were merged in [u/xp][upbound-xp-fork] and [u/uxp][uxp].
  - [ ] mention the [xp/xp][upstream-xp] version it refers to.
  - [ ] list new contributors to [u/uxp][uxp].
  - [ ] have the links to the full changelog of [u/xp][upbound-xp-fork] and [u/uxp][uxp].
- [ ] Ensured that users have been notified of the release on all communitcation channels:
  - [ ] Slack: crossposting on Crossplane's Slack workspace channels `#announcements`, `#upbound` and `#squad-crossplane` on Upbound's Slack.
  - [ ] Twitter: ask `#marketing` on Upbound's Slack to do so.


<!-- Named Links -->
[ci-uxp]: https://github.com/upbound/universal-crossplane/actions/workflows/ci.yml
[ci-xp-fork]: https://github.com/upbound/crossplane/actions/workflows/ci.yml
[promote-uxp]: https://github.com/upbound/universal-crossplane/actions/workflows/promote.yml
[sync-xp-fork]: https://github.com/upbound/universal-crossplane/blob/main/CONTRIBUTING.md#crossplane-fork-sync
[tag-uxp]: https://github.com/upbound/universal-crossplane/actions/workflows/tag.yml
[tag-xp-fork]: https://github.com/upbound/crossplane/actions/workflows/tag.yml
[upbound-xp-fork]: https://github.com/upbound/crossplane
[upstream-helm-chart]: https://github.com/crossplane/crossplane/tree/master/cluster/charts/crossplane
[upstream-xp-values]: https://github.com/crossplane/crossplane/blob/master/cluster/charts/crossplane/values.yaml.tmpl
[upstream-xp]: https://github.com/crossplane/crossplane
[uxp-main-channel]: https://charts.upbound.io/main
[uxp-releases]: https://github.com/upbound/universal-crossplane/releases
[uxp-schedule]: https://github.com/upbound/universal-crossplane/blob/main/README.md#releases
[uxp-stable-channel]: https://charts.upbound.io/stable
[uxp-values]: https://github.com/upbound/universal-crossplane/blob/main/cluster/charts/universal-crossplane/values.yaml.tmpl
[uxp]: https://github.com/upbound/universal-crossplane
