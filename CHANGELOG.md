# Changelog

## [0.1.7](https://github.com/leg100/pug/compare/v0.1.6...v0.1.7) (2024-05-02)


### Features

* add stale state to runs ([246315c](https://github.com/leg100/pug/commit/246315cf252fd24ea8d97ea79b92e2913174cb2f))
* move plan files to ~/.pug, and auto-delete them ([05c039d](https://github.com/leg100/pug/commit/05c039d71c1c2c826149e0896572dfd0a65412fd))


### Bug Fixes

* show '+0~0-0' when no changes, not '-' ([9661610](https://github.com/leg100/pug/commit/9661610d889d4afd71b4007e67249d906bc87c0c))
* terminate running tasks upon exit ([7ad8289](https://github.com/leg100/pug/commit/7ad8289bee4aeff5007058621e9e4b97d9f98e0f))


### Miscellaneous

* clean up pug path code ([81c27c1](https://github.com/leg100/pug/commit/81c27c1b8adab5ea9b689ad34b9f9a33768a6468))
* document resource hierarchy ([05e4932](https://github.com/leg100/pug/commit/05e4932a0cafb90279d0bfcb7f7e1affbe666b0f))
* git ignore asdf .tool-versions files ([3f598bb](https://github.com/leg100/pug/commit/3f598bbea52165beae0298ce4acc0f6583553d62))
* remove vhs vids from git ([2126a1b](https://github.com/leg100/pug/commit/2126a1b35821da53f3990a47f80d153de678359e))

## [0.1.6](https://github.com/leg100/pug/compare/v0.1.5...v0.1.6) (2024-04-30)


### Features

* auto load workspace tfvars file ([#45](https://github.com/leg100/pug/issues/45)) ([95ebc7e](https://github.com/leg100/pug/commit/95ebc7e4ee9d7e6c9aa8fe7eb4eba7a3ec89f08d))

## [0.1.5](https://github.com/leg100/pug/compare/v0.1.4...v0.1.5) (2024-04-30)


### Bug Fixes

* add user's env to tasks ([f71fd11](https://github.com/leg100/pug/commit/f71fd115e2a8e15935e02c2b408146755c6fc438))


### Miscellaneous

* add example to getting started ([e2c3658](https://github.com/leg100/pug/commit/e2c3658714b25ab39150d7b856cfcb2de9c6e26c))
* add getting started section ([98323e0](https://github.com/leg100/pug/commit/98323e0ff6ff0c192915887545e8accfbe43dab2))
* add start and finish msgs to demos ([fad6742](https://github.com/leg100/pug/commit/fad674205fe4a802304074a2dc97f12f90814b7e))
* update automatic tasks ([64552f7](https://github.com/leg100/pug/commit/64552f73f3f85cb7371dfffb4fa4a970c3694807))
* update getting started guide ([faf691a](https://github.com/leg100/pug/commit/faf691aabc55055c6de1f0f1e7ea4114785472e7))
* update readme.md ([49a71ed](https://github.com/leg100/pug/commit/49a71ed4f11d0b88705eca915e547db371863e81))

## [0.1.4](https://github.com/leg100/pug/compare/v0.1.3...v0.1.4) (2024-04-29)


### Features

* add support for destroy plan ([7dd5f9d](https://github.com/leg100/pug/commit/7dd5f9d28a1ac27ecfb15235019a3367ca0d5c30))
* press 'C' to change workspace ([9fca858](https://github.com/leg100/pug/commit/9fca85853180b7ff3fd0eb7ab5d548230859f7b0))
* prune selections prior to plan/apply ([#42](https://github.com/leg100/pug/issues/42)) ([b2b9902](https://github.com/leg100/pug/commit/b2b990206add8f0d42316787da3e22e8460afbcd))


### Bug Fixes

* add help binding for targeted plan ([f212080](https://github.com/leg100/pug/commit/f212080dda86e05a0638238669f3e1ecf52656d5))
* add missing navigation keys ([55d2b39](https://github.com/leg100/pug/commit/55d2b397e066a54a27f24aef8ef687eb4e7f46ec))
* de-select rows after triggering plans ([c4ff893](https://github.com/leg100/pug/commit/c4ff89330e825babed3a6c25bcb0c8a318b555cb))
* don't unsub full subs ([7215a79](https://github.com/leg100/pug/commit/7215a79bcd17ed96d8d25ef925cbcb19c56e89c9))
* flaky tests ([eff8924](https://github.com/leg100/pug/commit/eff8924abf60c47a572a5d35e89fe7de3c01b666))
* flaky workspace destroy test ([126e35d](https://github.com/leg100/pug/commit/126e35d94a212bb6f83fc4aeb28215d5275789c1))
* main image on README 404 ([b63cad4](https://github.com/leg100/pug/commit/b63cad4f23591cb16664ba778ee19a53cf37bba9))
* pubsub broker now can handle unlimited events ([9f55036](https://github.com/leg100/pug/commit/9f550361bbf7e57deb6ccfd3ed69f4bdc6a277e9))
* workspace list test missing workspace fixture ([294765f](https://github.com/leg100/pug/commit/294765f59f746e67908930b305091364001588d2))


### Miscellaneous

* add tasks demo ([1cb2969](https://github.com/leg100/pug/commit/1cb29693a98e60e3353380adcf3a45de322d00ef))
* add test for reload workspaces ([e5fd198](https://github.com/leg100/pug/commit/e5fd198456cc4367d672bdace11a90418338be22))
* bug report issue template ([275a47c](https://github.com/leg100/pug/commit/275a47cd7662e64e3c085a543cb25e7943a1cfbc))
* feature request issue template ([b6ea793](https://github.com/leg100/pug/commit/b6ea793092c5b139c390a01e6e6eef63e5265221))
* provide info on module reload ([201f913](https://github.com/leg100/pug/commit/201f9131f33329c2d24a95c664c5fa5ace54349b))
* remove t.Log ([d550715](https://github.com/leg100/pug/commit/d550715811292b280730a1acf8bf7b0707a2d4f7))
* remove unused pre-commit config ([033cefe](https://github.com/leg100/pug/commit/033cefe080655183c55a79684ab9bb93cdbc7f29))
* update demos ([d9d40a1](https://github.com/leg100/pug/commit/d9d40a13d7db47463aa5c052e037a70849498bfd))

## [0.1.3](https://github.com/leg100/pug/compare/v0.1.2...v0.1.3) (2024-04-24)


### Bug Fixes

* brew release missing token env ([751f6f1](https://github.com/leg100/pug/commit/751f6f18876dc3c2a1ec37e712c57bf437253258))

## [0.1.2](https://github.com/leg100/pug/compare/v0.1.1...v0.1.2) (2024-04-24)


### Bug Fixes

* surface version ([6c6b086](https://github.com/leg100/pug/commit/6c6b086c1886c69b5f42ec07779b4c001a497e08))


### Miscellaneous

* create brew tap ([81ddc8f](https://github.com/leg100/pug/commit/81ddc8fa789a55ae122eb536752ebab2ea9a5726))

## [0.1.1](https://github.com/leg100/pug/compare/v0.1.0...v0.1.1) (2024-04-24)


### Bug Fixes

* remove unnecessary code failing win build ([07426c5](https://github.com/leg100/pug/commit/07426c5df914c0a0300efe16c3edd8ff89b8f9e3))

## [0.1.0](https://github.com/leg100/pug/compare/v0.0.1...v0.1.0) (2024-04-24)


### ⚠ BREAKING CHANGES

* initial commit

### Features

* initial commit ([25436c4](https://github.com/leg100/pug/commit/25436c4d4a2e2a75363824b5f2ce27815b7f1079))
