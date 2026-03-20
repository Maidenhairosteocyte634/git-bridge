# Changelog

All notable changes to this project will be documented in this file.

## [v0.3.0](https://github.com/somaz94/git-bridge/compare/v0.2.0...v0.3.0) (2026-03-20)

### Features

- sync with internal repo - commit author, porcelain push, config validation ([5b0d904](https://github.com/somaz94/git-bridge/commit/5b0d904f996df8121e82da7a696fba4b64f0df76))
- add CODEOWNERS ([1ad31e0](https://github.com/somaz94/git-bridge/commit/1ad31e068d0d3d01847db9511155ebfb0b84bdc4))

### Bug Fixes

- use GITHUB_TOKEN for dependabot auto merge ([83c06dd](https://github.com/somaz94/git-bridge/commit/83c06dd83b383d48046d68a43d31c0e19a7cb87e))

### Documentation

- add no-push rule to CLAUDE.md ([6e640ae](https://github.com/somaz94/git-bridge/commit/6e640ae398829a344eda10b8c2ea7aab8f73f872))
- update CONTRIBUTORS.md ([3e0b3db](https://github.com/somaz94/git-bridge/commit/3e0b3db2df82d40498be5f663f9862c27d29bb31))
- update changelog ([48294b6](https://github.com/somaz94/git-bridge/commit/48294b69e6295abdcb72c0f7cfda951a0239c838))

### Builds

- **deps:** Bump the go-minor group with 4 updates (#2) ([#2](https://github.com/somaz94/git-bridge/pull/2)) ([7fda2df](https://github.com/somaz94/git-bridge/commit/7fda2dfe2ac3db9454095159089d08445229f297))

### Continuous Integration

- migrate gitlab-mirror workflow to multi-git-mirror action ([3ab4ce3](https://github.com/somaz94/git-bridge/commit/3ab4ce3f46d1ecf375d8dbd70179794270a3cdb5))

### Contributors

- somaz

<br/>

## [v0.2.0](https://github.com/somaz94/git-bridge/compare/v0.1.0...v0.2.0) (2026-03-18)

### Features

- add committer info (Pushed by) to Slack mirror sync notifications ([36acacd](https://github.com/somaz94/git-bridge/commit/36acacdf72dc83e6b60acb147a05856fbefc9b96))
- implement incremental fetch with PVC-backed mirror cache ([5c402e5](https://github.com/somaz94/git-bridge/commit/5c402e59ac39cc4ee8f380b6b798fd9df25c32b8))

### Documentation

- CLAUDE.md ([0acbff3](https://github.com/somaz94/git-bridge/commit/0acbff35a74d5ac485572df8a9e21a44102b5bb9))
- add CLAUDE.md project guide ([1afd0ef](https://github.com/somaz94/git-bridge/commit/1afd0ef3f3a10db20cbb55fa406103530eb8748c))
- update changelog ([daec38c](https://github.com/somaz94/git-bridge/commit/daec38c13e94db0b2bbafcde3495976a9cc76f58))
- README.md ([2064c8a](https://github.com/somaz94/git-bridge/commit/2064c8ad404f2b3029a07652eecb44cdc6ca7aa9))
- update CONTRIBUTORS.md ([20383cd](https://github.com/somaz94/git-bridge/commit/20383cdb74a7989122dedca2247652fe0b320f32))
- update changelog ([0104baa](https://github.com/somaz94/git-bridge/commit/0104baa44fdac57ec6136504264e516dcae853ac))

### Tests

- improve coverage from 93% to 97.9% and separate make test/test-cover roles ([0f65504](https://github.com/somaz94/git-bridge/commit/0f65504f047a4132e9312095aa64fd49b788ed5c))

### Continuous Integration

- use somaz94/contributors-action@v1 for contributors generation ([49fd3a5](https://github.com/somaz94/git-bridge/commit/49fd3a56852728eb8b5eb35ea6954d156e916803))
- use major-tag-action for version tag updates ([11b9d93](https://github.com/somaz94/git-bridge/commit/11b9d9356498ab84e53301ce1ddccb0ea81504cf))
- migrate changelog generator to go-changelog-action ([6510563](https://github.com/somaz94/git-bridge/commit/65105638df73f3ea8139b396c40470e07fc8efe3))
- add GitHub release notes configuration ([4fbc5d9](https://github.com/somaz94/git-bridge/commit/4fbc5d95d0693f94680bf77e4a39b5485f9c5eff))
- unify changelog-generator with flexible tag pattern ([a8778f6](https://github.com/somaz94/git-bridge/commit/a8778f6ceed28908975c22cea9fb8b285ccd5574))

### Contributors

- somaz

<br/>

## [v0.1.0](https://github.com/somaz94/git-bridge/compare/v0.0.1...v0.1.0) (2026-03-13)

### Features

- add DockerHub multi-arch build and push support ([2c0aca7](https://github.com/somaz94/git-bridge/commit/2c0aca7c709ce510aa4a0000dcba1ab85c612218))
- add K8s manifests and example configurations ([b25c610](https://github.com/somaz94/git-bridge/commit/b25c610480486088e0ce77d9cb1a96a2144784b4))
- add core mirroring engine with multi-provider support ([f70823e](https://github.com/somaz94/git-bridge/commit/f70823ef46b6fd4712815e93e97a8f05d5f1d912))

### Bug Fixes

- skip major version tag deletion on first release ([cbadec1](https://github.com/somaz94/git-bridge/commit/cbadec148ae7e35b1560c74cca85b6721ce7fd5c))
- remove docker job from release workflow ([580e593](https://github.com/somaz94/git-bridge/commit/580e593305e20a4c1af308c007343d3a5064a1c3))
- fix changelog-generator tag handling and dependabot secrets access ([553a875](https://github.com/somaz94/git-bridge/commit/553a875849cd975aa45e74986c78f77ce58e3166))

### Documentation

- add documentation, architecture diagram, and update README ([c4d3418](https://github.com/somaz94/git-bridge/commit/c4d341832f629610a0fd4760a5870cf24d751432))
- add documentation, architecture diagram, and update README ([3154eae](https://github.com/somaz94/git-bridge/commit/3154eae3430e8ff0d810e3fb2b7b6d3db4033630))

### Builds

- **deps:** Bump alpine from 3.21 to 3.23 in the docker-minor group ([cb2d032](https://github.com/somaz94/git-bridge/commit/cb2d03215e8b9b0dac690817fb5b4a4b63700e8f))
- **deps:** Bump alpine from 3.21 to 3.23 in the docker-minor group ([1e387db](https://github.com/somaz94/git-bridge/commit/1e387dba39f0c50de032915e88a0ed2d1189f123))

### Continuous Integration

- add GitHub Actions workflows and dependabot config ([a73d969](https://github.com/somaz94/git-bridge/commit/a73d9699f6f14bd08d26b2f8d8a0c7be30785df0))

### Contributors

- somaz

<br/>

## [v0.0.1](https://github.com/somaz94/git-bridge/releases/tag/v0.0.1) (2026-03-13)

### Contributors

- somaz

<br/>

