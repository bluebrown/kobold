# Changelog

## [0.4.2](https://github.com/bluebrown/kobold/compare/v0.4.1...v0.4.2) (2024-12-12)


### Bug Fixes

* Main commit message ([#72](https://github.com/bluebrown/kobold/issues/72)) ([eb1cfc6](https://github.com/bluebrown/kobold/commit/eb1cfc6b1f4c42ae81429b84830ceeaf0c52fe0d))

## [0.4.1](https://github.com/bluebrown/kobold/compare/v0.4.0...v0.4.1) (2024-10-09)


### Features

* Improve commit messages to include change detail ([2f43914](https://github.com/bluebrown/kobold/commit/2f43914344465fffcded4849aa0d212f91b68946))


### Bug Fixes

* handle commit message format errors ([#69](https://github.com/bluebrown/kobold/issues/69)) ([1971e78](https://github.com/bluebrown/kobold/commit/1971e784126c2d8e0fa1c42b184cecef559f2444))

## [0.4.0](https://github.com/bluebrown/kobold/compare/v0.3.3...v0.4.0) (2024-08-26)


### ⚠ BREAKING CHANGES

* dont strip .git suffix for interal repo string. may break custom posthooks
* **pkguri:** Parse ref and pkg as query params ([#62](https://github.com/bluebrown/kobold/issues/62))

### Features

* confix for new pkg uri format ([f158a7e](https://github.com/bluebrown/kobold/commit/f158a7e22552f283bc0ead9a22c1251b02c18c99))


### Bug Fixes

* pullrequest posthook message ([#65](https://github.com/bluebrown/kobold/issues/65)) ([71283eb](https://github.com/bluebrown/kobold/commit/71283eb93fea29af08c5226e9f98574f6b3df8a7))


### Code Refactoring

* dont strip .git suffix for interal repo string. may break custom posthooks ([f158a7e](https://github.com/bluebrown/kobold/commit/f158a7e22552f283bc0ead9a22c1251b02c18c99))
* **pkguri:** Parse ref and pkg as query params ([#62](https://github.com/bluebrown/kobold/issues/62)) ([e6ad83e](https://github.com/bluebrown/kobold/commit/e6ad83e4f35830844db979b989f23f23b1c8f89d))

## [0.3.3](https://github.com/bluebrown/kobold/compare/v0.3.2...v0.3.3) (2024-06-03)


### Bug Fixes

* **filter:** false positive change reports due to context ([#58](https://github.com/bluebrown/kobold/issues/58)) ([e787d10](https://github.com/bluebrown/kobold/commit/e787d104990cba75551390b683031bb3263c1d56))

## [0.3.2](https://github.com/bluebrown/kobold/compare/v0.3.1...v0.3.2) (2024-05-28)


### Features

* Apple Silicon local development support ([#50](https://github.com/bluebrown/kobold/issues/50)) ([ab14a59](https://github.com/bluebrown/kobold/commit/ab14a59f9d194cbe310453bec076ae89d9004079))
* **filter:** handle 'tag only' fields ([#53](https://github.com/bluebrown/kobold/issues/53)) ([7ae586e](https://github.com/bluebrown/kobold/commit/7ae586e00bf45a72d3cf28fb2fe23690ff7b0143))
* **filter:** handle digest based parts ([#57](https://github.com/bluebrown/kobold/issues/57)) ([066dcfe](https://github.com/bluebrown/kobold/commit/066dcfeea3635e41472b780200d6b7d6c294e514))
* Include updated images in commit message ([#49](https://github.com/bluebrown/kobold/issues/49)) ([aab019c](https://github.com/bluebrown/kobold/commit/aab019c3f861e285e3c6f9465bab8ac240eb6f78))

## [0.3.1](https://github.com/bluebrown/kobold/compare/v0.3.0...v0.3.1) (2024-04-14)


### Features

* harbor decoder ([979090f](https://github.com/bluebrown/kobold/commit/979090fe689c0f84fda2d9cfd20cc381304aba2d))
* webhook channel query param ([#40](https://github.com/bluebrown/kobold/issues/40)) ([806fde4](https://github.com/bluebrown/kobold/commit/806fde47ce183edfb09abf59da0fc2fb9fa8b6b2))

## [0.3.0](https://github.com/bluebrown/kobold/compare/v0.2.4...v0.3.0) (2024-01-25)


### ⚠ BREAKING CHANGES

* complete rewrite ([#36](https://github.com/bluebrown/kobold/issues/36))

### Code Refactoring

* complete rewrite ([#36](https://github.com/bluebrown/kobold/issues/36)) ([b178a55](https://github.com/bluebrown/kobold/commit/b178a5577436d04d6a644476426eb7ec6fe975f1))

## Changelog
