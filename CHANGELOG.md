# Changelog

## [0.3.1](https://github.com/leg100/pug/compare/v0.3.0...v0.3.1) (2024-07-02)


### Miscellaneous

* make colors somewhat less garish ([ad2bdff](https://github.com/leg100/pug/commit/ad2bdff792879b55dd7cc9453d2d7129cc8bac73))

## [0.3.0](https://github.com/leg100/pug/compare/v0.2.2...v0.3.0) (2024-07-01)


### ⚠ BREAKING CHANGES

* change key bindings for split pane resizing ([#82](https://github.com/leg100/pug/issues/82))
* change default config file path ([#79](https://github.com/leg100/pug/issues/79))

### refactor

* change default config file path ([#79](https://github.com/leg100/pug/issues/79)) ([d1d9b4e](https://github.com/leg100/pug/commit/d1d9b4ef0c8112ae57f7b7b250c8cc9ed0c9666c))
* change key bindings for split pane resizing ([#82](https://github.com/leg100/pug/issues/82)) ([35f33d1](https://github.com/leg100/pug/commit/35f33d16786afc258ba37a642be02cc68aaa8841))


### Bug Fixes

* always ensure current row is visible ([e4ddc3a](https://github.com/leg100/pug/commit/e4ddc3a7791e1b44ad5df0684b0a2e68d5d523ac))
* remove debug table info ([916a859](https://github.com/leg100/pug/commit/916a8591b04d3c5fa4901d444ed62b73e3545e8e))
* use terraform for terragrunt tests ([e0cd4b3](https://github.com/leg100/pug/commit/e0cd4b38ae844c79fb6be0cfbd16698125f36f90))


### Miscellaneous

* update demo ([348c3b5](https://github.com/leg100/pug/commit/348c3b5fb5334ea12e2af88dfab16fee56ad9cfc))

## [0.2.2](https://github.com/leg100/pug/compare/v0.2.1...v0.2.2) (2024-06-23)


### Features

* terragrunt mode ([#77](https://github.com/leg100/pug/issues/77)) ([9be2914](https://github.com/leg100/pug/commit/9be29144fef6c1ed9c810f1393c39f44949cc06e))


### Bug Fixes

* border w/o preview nr invisible on light bg ([ba2313a](https://github.com/leg100/pug/commit/ba2313a7a7d76721309805e2d7e417c8ac01901a))
* detect applies with no changes ([84cfb6f](https://github.com/leg100/pug/commit/84cfb6f323d75f02b915e96f50e29199f012415b))


### Miscellaneous

* remove run status from UI ([805eb6b](https://github.com/leg100/pug/commit/805eb6bc2e54a68592d4ce9ec283037a782e9bfb))
* update tofu/terragrunt support docs ([394f173](https://github.com/leg100/pug/commit/394f173e0c72265c79f7cc0895343057cecac945))

## [0.2.1](https://github.com/leg100/pug/compare/v0.2.0...v0.2.1) (2024-06-21)


### Features

* require approval before retrying tasks ([0f5e7e3](https://github.com/leg100/pug/commit/0f5e7e38cdf74fe6c2a5b2f55aaa9f6c46529ecf))


### Bug Fixes

* go install broken by replace directive ([8ab6fb3](https://github.com/leg100/pug/commit/8ab6fb35b8be4244edbaa990ea8e221bc492bbf9))
* provide further info when pruning selection ([4c863e9](https://github.com/leg100/pug/commit/4c863e91114ec6b169d684d4e059fc2a5582f4f1))
* table current row always track item ([e9c673b](https://github.com/leg100/pug/commit/e9c673b9bbb1b4bcb01dae68507ce727a6ce02a7))


### Miscellaneous

* add terminal trove badge to README.md ([3b0c0a2](https://github.com/leg100/pug/commit/3b0c0a2c5db7714165f8e6f4953d46453d3e20e5))
* removed unused progress bar ([f5fa390](https://github.com/leg100/pug/commit/f5fa3909b880982e95f80a330db0e40210180441))
* styling changes ([51ab14b](https://github.com/leg100/pug/commit/51ab14b241b8a1b556bb238c4db2255b2ba588ce))
* update demo ([dbb8458](https://github.com/leg100/pug/commit/dbb845870cde3fc1c518805581119994962bc482))

## [0.2.0](https://github.com/leg100/pug/compare/v0.1.11...v0.2.0) (2024-06-20)


### ⚠ BREAKING CHANGES

* bump minor version

### Features

* add commands to state resource page ([9550973](https://github.com/leg100/pug/commit/955097341ec3d690e9b812da6524e443a8945b45))
* add more task info ([26d25c8](https://github.com/leg100/pug/commit/26d25c854c5b42e23ed60cd8f53cb18090adbf51))
* bump minor version ([2d0bff0](https://github.com/leg100/pug/commit/2d0bff08f2ce8ad92bf04d644b61ddac2d24ab97))
* retry tasks ([e5ef0c4](https://github.com/leg100/pug/commit/e5ef0c4f447bbaddcfabd320f1f44e7805d7ec7d))
* show error when key is unknwon ([e07c455](https://github.com/leg100/pug/commit/e07c4551d250ad095ed3170718d411a0d405de61))
* show spinner when waiting for output ([bda279d](https://github.com/leg100/pug/commit/bda279d448ae0880c1497d7509aeac88a8950064))
* split state page ([bfb86f4](https://github.com/leg100/pug/commit/bfb86f46a2ab4c842ab8c9d3758caf1c6169d3aa))
* task groups and split screen  ([#69](https://github.com/leg100/pug/issues/69)) ([869a790](https://github.com/leg100/pug/commit/869a7901ddd26c8ace015397e8e416fc81f4001f))
* task info sidebar ([a888922](https://github.com/leg100/pug/commit/a8889226d5e12c9047ebe7e13d17cfe97f5cdd77))
* toggle autoscroll ([57234fd](https://github.com/leg100/pug/commit/57234fd42b4482ddeb480917ae6446473b85b971))


### Bug Fixes

* adding missing key bindings to help ([368d59e](https://github.com/leg100/pug/commit/368d59ea79c5e63f3584e0c95604d66e7fc5d716))
* consistenly format error messages ([f4f02e0](https://github.com/leg100/pug/commit/f4f02e005c5abf4ad861e9197cbf17c21685e3a0))
* get tests passing again ([#72](https://github.com/leg100/pug/issues/72)) ([fa482ce](https://github.com/leg100/pug/commit/fa482ce52c8dca68fb49dca88b608c0380ebf4af))
* handle empty state without panic ([26028c2](https://github.com/leg100/pug/commit/26028c253d516e1ea991da86b67fd3fdeb345f38))
* handling unknown keys is difficult so remove err ([eea8b49](https://github.com/leg100/pug/commit/eea8b49be5ee1a7a1a97ec67eab62fc73228e034))
* integration tests use mirror ([a5a0f5a](https://github.com/leg100/pug/commit/a5a0f5ac1407967382d616aa5784c59c427def05))
* keep track of cursor on table ([bd4b8eb](https://github.com/leg100/pug/commit/bd4b8ebb618527fe86beeb2bbea4676091aa46b9))
* key changes broke tests ([866b0d2](https://github.com/leg100/pug/commit/866b0d2d6f3025de3e7250ebfc38e808cb19c219))
* remove double border on task group table ([2fe00a6](https://github.com/leg100/pug/commit/2fe00a6e39bad0b2a5698185e3c04c46b63dcf0f))
* show task's workspace/module in logs ([4a67df5](https://github.com/leg100/pug/commit/4a67df5e977aaf48c4b045abab4d935c3d6e34d3))
* use absolute path for edit ([8c0aac1](https://github.com/leg100/pug/commit/8c0aac12f11378bbe560a36be3c768e760876431))
* use EDITOR for editing modules ([d4bb1f4](https://github.com/leg100/pug/commit/d4bb1f4d6f3d58e4b59aa1643b7bbbce0d47bf17))


### Miscellaneous

* bump go deps ([7f2e8d2](https://github.com/leg100/pug/commit/7f2e8d2b47ee37277d953cfd2a2cf730b481db98))
* change cancel language ([6d90e7e](https://github.com/leg100/pug/commit/6d90e7ed10b13d7eebf51339a222fc96f71324b9))
* clean up table naming ([c3cdf2e](https://github.com/leg100/pug/commit/c3cdf2e43821794cc9c20f984fc384e4c2fcc3e5))
* enable autoscroll by default ([62f95ae](https://github.com/leg100/pug/commit/62f95ae33ae2aac83ecf86093e308145a35dd194))
* lint changes ([72820ab](https://github.com/leg100/pug/commit/72820ab2e3c050266c83f275bcb83141cb7ddf3d))
* merge demos into one and update readme ([#73](https://github.com/leg100/pug/issues/73)) ([f2cf271](https://github.com/leg100/pug/commit/f2cf271e118590323de5b5780a8e2aec71e0ae26))
* merge table types into one ([#68](https://github.com/leg100/pug/issues/68)) ([6bcc5fb](https://github.com/leg100/pug/commit/6bcc5fb898629a89b9c44017cb3d5890a3f97a68))
* refactor resources ([#64](https://github.com/leg100/pug/issues/64)) ([16741d7](https://github.com/leg100/pug/commit/16741d71150c3a1af19df60855d34d7ea96a1867))
* refactor state ([#67](https://github.com/leg100/pug/issues/67)) ([8f516c4](https://github.com/leg100/pug/commit/8f516c43c51b197a0d224f82c06bfb578784fa94))
* remove leftover tab code ([b60fc0d](https://github.com/leg100/pug/commit/b60fc0d91bbf4b6fb891e024a099c80c3394cbb6))
* remove redundant run key ([382f9fe](https://github.com/leg100/pug/commit/382f9fefa63f5d3143c59acc0038f6c911a0cc37))
* remove redundant table.Items() method ([5bc749d](https://github.com/leg100/pug/commit/5bc749da374e851e376250bd68a89c2b0ddbd685))
* remove shift-tab key binding to go back ([e5f4e11](https://github.com/leg100/pug/commit/e5f4e1146e1ea32204d1444a60e86fa0b5efdb70))
* remove unnecessary generic table param ([3871c1f](https://github.com/leg100/pug/commit/3871c1fba50a9b31896ccce540fb481038916431))
* rm unnecess receiver from crumbs method ([5d32a7a](https://github.com/leg100/pug/commit/5d32a7a6ef035530ad73660c2b7bd5b6229090f0))
* rm unnecessary mod retrieval ([1f61982](https://github.com/leg100/pug/commit/1f619827890194908d48258ff004ae4a4535579c))

## [0.1.11](https://github.com/leg100/pug/compare/v0.1.10...v0.1.11) (2024-05-13)


### Features

* consistent navigation following task/run creation ([#60](https://github.com/leg100/pug/issues/60)) ([bbf97cd](https://github.com/leg100/pug/commit/bbf97cdd14c52bc6981cddabfb5aea1dbd681924))


### Bug Fixes

* go install does not like replace directives ([#63](https://github.com/leg100/pug/issues/63)) ([55ab62a](https://github.com/leg100/pug/commit/55ab62a4eebb2c7f43a37448aab55c08644b6338))
* show version when using go install ([eb27209](https://github.com/leg100/pug/commit/eb27209f7d0660504693961587aa44c74becbb0c))

## [0.1.10](https://github.com/leg100/pug/compare/v0.1.9...v0.1.10) (2024-05-10)


### Features

* add ability to move state resources ([#57](https://github.com/leg100/pug/issues/57)) ([0945bb8](https://github.com/leg100/pug/commit/0945bb81a14b80a1915ae307964cb23fba336a8c))
* add resource count to module listing ([b1bf226](https://github.com/leg100/pug/commit/b1bf226ee4126a4ecf5f3adabf6ff53cb0efd66d))


### Bug Fixes

* make select range behave like k9 ([6e38d37](https://github.com/leg100/pug/commit/6e38d37468b4ebdf759c4b17e6add990287b91a3))
* state reload not always visible ([942943b](https://github.com/leg100/pug/commit/942943bd0b3d5c0207c510c7626de167ec9e1be6))


### Miscellaneous

* remove ` as back key ([1deaec1](https://github.com/leg100/pug/commit/1deaec1180b663e60a52a7e89aae492c0e0410f7))
* remove unnecessary test setup options ([b3fb7be](https://github.com/leg100/pug/commit/b3fb7bebbd512ce101340243715835c805c8ae0e))
* remove unnecessary update in reload ([9ae18cf](https://github.com/leg100/pug/commit/9ae18cf651d1401f72afa57e9dce2b2df824e395))

## [0.1.9](https://github.com/leg100/pug/compare/v0.1.8...v0.1.9) (2024-05-07)


### Features

* add keybinding &lt;shift&gt;+<tab> to go back ([7bb8c2b](https://github.com/leg100/pug/commit/7bb8c2b6f8181361c5f51831788b189a783b4806))
* drill down into log message ([#56](https://github.com/leg100/pug/issues/56)) ([ceb5559](https://github.com/leg100/pug/commit/ceb55590f0914a08ac5be828e446d8f999ffa224))
* filter table rows ([#55](https://github.com/leg100/pug/issues/55)) ([7b89b70](https://github.com/leg100/pug/commit/7b89b70769264cc33a6156cfe25c626d95c22ff7))


### Bug Fixes

* make help bindings consistent ([d06d8e0](https://github.com/leg100/pug/commit/d06d8e01ad42f750c183f2898238d37e80afc2b4))
* prevent broker deadlock upon shutdown ([ebcb4d2](https://github.com/leg100/pug/commit/ebcb4d28a339448b473e7085099e6c65c3a01c17))


### Miscellaneous

* parallelize tests ([#54](https://github.com/leg100/pug/issues/54)) ([5ca1dd3](https://github.com/leg100/pug/commit/5ca1dd3b40df41f564d953f4bfee7b8a4a745e00))

## [0.1.8](https://github.com/leg100/pug/compare/v0.1.7...v0.1.8) (2024-05-03)


### Features

* don't auto-deselect and add select-range ([fa4eeee](https://github.com/leg100/pug/commit/fa4eeeef4e68e1d380f9c7d4250d250359e5b998))


### Bug Fixes

* comment out demo welcome message ([7653899](https://github.com/leg100/pug/commit/765389966bc17950901e625457ecf3bb283df389))
* hide apply key for non-applyable runs ([82f6880](https://github.com/leg100/pug/commit/82f6880c5c8b4b3e6f2e918ef38e7b5595ec7375))
* tab info now uses active tab ([fbed323](https://github.com/leg100/pug/commit/fbed323d1d5a750743a574c6a6d81fd90d482c2b))

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
