# Changelog

## [1.7.0](https://github.com/bucketeer-io/go-server-sdk/compare/v1.6.1...v1.7.0) (2026-02-06)


### Features

* add retry configuration options for http requests ([#183](https://github.com/bucketeer-io/go-server-sdk/issues/183)) ([c9d32bb](https://github.com/bucketeer-io/go-server-sdk/commit/c9d32bbb8f630b568dc5798bddd63ceb6f849978))
* add event stats to expose event processing metrics ([#189](https://github.com/bucketeer-io/go-server-sdk/issues/189)) ([4d2c23a](https://github.com/bucketeer-io/go-server-sdk/commit/4d2c23a8c60a8a71a28e20d6ea4ddc19b89b1d07))

### Improvements

* replace channel queue with ring buffer for improved performance ([#189](https://github.com/bucketeer-io/go-server-sdk/issues/189)) ([4d2c23a](https://github.com/bucketeer-io/go-server-sdk/commit/4d2c23a8c60a8a71a28e20d6ea4ddc19b89b1d07))


### Bug Fixes

* archived flag not being deleted from local cache ([#181](https://github.com/bucketeer-io/go-server-sdk/issues/181)) ([0b08c02](https://github.com/bucketeer-io/go-server-sdk/commit/0b08c022487a41800777bed7ffd36612f6f5e110))
* classify errors by type for metrics instead of string matching ([#188](https://github.com/bucketeer-io/go-server-sdk/issues/188)) ([786f0bf](https://github.com/bucketeer-io/go-server-sdk/commit/786f0bf44c819ac5b6cb33e83983058024caf197))


### Miscellaneous

* optimize http timeout and flush interval for server-side usage ([#190](https://github.com/bucketeer-io/go-server-sdk/issues/190)) ([0a8df9b](https://github.com/bucketeer-io/go-server-sdk/commit/0a8df9b2affec4f351fbe925313812a63025be3c))
* changed default event flush interval to 30s

## [1.6.1](https://github.com/bucketeer-io/go-server-sdk/compare/v1.6.0...v1.6.1) (2026-01-20)


### Bug Fixes

* add semver support for equals and in operators with v-prefix normalization ([#178](https://github.com/bucketeer-io/go-server-sdk/issues/178)) ([905d292](https://github.com/bucketeer-io/go-server-sdk/commit/905d292470aabb9a1b3ef9bc193030438e45348b))

## [1.6.0](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.5...v1.6.0) (2025-09-26)


### Features

* add error details reason types ([#167](https://github.com/bucketeer-io/go-server-sdk/issues/167)) ([c7ec90b](https://github.com/bucketeer-io/go-server-sdk/commit/c7ec90b76d1397a67345a29d1251d00951f706e5))
* add scheme configuration ([#159](https://github.com/bucketeer-io/go-server-sdk/issues/159)) ([3ca01d3](https://github.com/bucketeer-io/go-server-sdk/commit/3ca01d3255818811c781769298bd2f0f97f06c36))
* add SDK version support to options ([#166](https://github.com/bucketeer-io/go-server-sdk/issues/166)) ([2655753](https://github.com/bucketeer-io/go-server-sdk/commit/2655753db6edba57a090071b81b82f5ac1b0356e))
* add source ID support to SDK options and evaluation requests ([#169](https://github.com/bucketeer-io/go-server-sdk/issues/169)) ([439c6da](https://github.com/bucketeer-io/go-server-sdk/commit/439c6da38f9e04b650cb89cb1cea69cac8f4828c))
* support detailed error reasons in client-side evaluation ([#174](https://github.com/bucketeer-io/go-server-sdk/issues/174)) ([7939bd8](https://github.com/bucketeer-io/go-server-sdk/commit/7939bd82f801bfdde0a74b0c257496ee23c1fd31))


### Bug Fixes

* remove unnecessary type arguments ([#163](https://github.com/bucketeer-io/go-server-sdk/issues/163)) ([11b5135](https://github.com/bucketeer-io/go-server-sdk/commit/11b51358c8d08af52a61b379049e0269d179278c))
* resolve missing dependencies in incremental feature flag evaluation ([#177](https://github.com/bucketeer-io/go-server-sdk/issues/177)) ([6c691c9](https://github.com/bucketeer-io/go-server-sdk/commit/6c691c9c6f53d11fd545a2ded6e8de2515089289))


### Miscellaneous

* deprecate host and port configs ([#162](https://github.com/bucketeer-io/go-server-sdk/issues/162)) ([6af938e](https://github.com/bucketeer-io/go-server-sdk/commit/6af938e0c5347b7d31e535bf75577df345f72db3))
* enhance SDK with versioning support in cache processors and requests ([#168](https://github.com/bucketeer-io/go-server-sdk/issues/168)) ([1fbdec5](https://github.com/bucketeer-io/go-server-sdk/commit/1fbdec50a4469a58c985dcea4bcf0ea6a226e753))
* fix CODEOWNERS ([#164](https://github.com/bucketeer-io/go-server-sdk/issues/164)) ([a068946](https://github.com/bucketeer-io/go-server-sdk/commit/a068946e5bbe74a67bb159bc1a75f55ca1f8a4e7))
* fix data race when calling IsReady() ([#175](https://github.com/bucketeer-io/go-server-sdk/issues/175)) ([ea1d748](https://github.com/bucketeer-io/go-server-sdk/commit/ea1d748fddf8ecd37e1294c6b291c24067d41090))
* rename WithSDKVersion to WithWrapperSDKVersion for internal use indication ([#170](https://github.com/bucketeer-io/go-server-sdk/issues/170)) ([9f9a207](https://github.com/bucketeer-io/go-server-sdk/commit/9f9a2074ef4988d4c6b14f20bc964fcbe46587a0))
* validate the source ID when initializing the SDK ([#173](https://github.com/bucketeer-io/go-server-sdk/issues/173)) ([9e1d5ff](https://github.com/bucketeer-io/go-server-sdk/commit/9e1d5ff17a1d9283a9134de122b88a539789402b))

## [1.5.5](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.4...v1.5.5) (2025-03-03)


### Bug Fixes

* flag not found when using dependency flag in targeting rule ([#157](https://github.com/bucketeer-io/go-server-sdk/issues/157)) ([623594d](https://github.com/bucketeer-io/go-server-sdk/commit/623594d663ccd175291cdb8dbca943a809b1755c))


### Miscellaneous

* chore: improve bucket hash algorithm using murmurHash3 instead of md5 (Bucketeer 1.3.0) ([#155](https://github.com/bucketeer-io/go-server-sdk/issues/155)) ([4af5f4c](https://github.com/bucketeer-io/go-server-sdk/commit/4af5f4c4eac6ebb60477ae40e6d8cd3d94a03da7))

## [1.5.4](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.3...v1.5.4) (2025-02-07)


### Bug Fixes

* unknow field error when changing the proto message on the backend ([#152](https://github.com/bucketeer-io/go-server-sdk/issues/152)) ([0373080](https://github.com/bucketeer-io/go-server-sdk/commit/0373080aa8cc84e339185681de324be274fe8d0d))


### Build System

* **deps:** update dependencies ([#153](https://github.com/bucketeer-io/go-server-sdk/issues/153)) ([40c1eeb](https://github.com/bucketeer-io/go-server-sdk/commit/40c1eebbbeb763664f7001466dda66584e0d8644))

## [1.5.3](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.2...v1.5.3) (2024-12-27)


### Bug Fixes

* evaluation event being reported with no user ID ([#149](https://github.com/bucketeer-io/go-server-sdk/issues/149)) ([948a5d3](https://github.com/bucketeer-io/go-server-sdk/commit/948a5d30139f12d3231e6ea61eafa50178d9dbc4))

## [1.5.2](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.1...v1.5.2) (2024-12-02)


### Build System

* **deps:** update bucketeer module to 1.1.0 ([#145](https://github.com/bucketeer-io/go-server-sdk/issues/145)) ([16bdba1](https://github.com/bucketeer-io/go-server-sdk/commit/16bdba1a89c42be6497cb585e8b0bcd1b3552873))

## [1.5.1](https://github.com/bucketeer-io/go-server-sdk/compare/v1.5.0...v1.5.1) (2024-10-18)


### Bug Fixes

* skip event generation for unauthorized and forbidden errors ([#143](https://github.com/bucketeer-io/go-server-sdk/issues/143)) ([f68875e](https://github.com/bucketeer-io/go-server-sdk/commit/f68875e6d62209f676521891a69f830c95272b54))

## [1.5.0](https://github.com/bucketeer-io/go-server-sdk/compare/v1.4.1...v1.5.0) (2024-09-17)


### Features

* get variation details by variation type ([#136](https://github.com/bucketeer-io/go-server-sdk/issues/136)) ([3fb2301](https://github.com/bucketeer-io/go-server-sdk/commit/3fb2301990ccf846d5e0a9a71e84d4e0439c4dbf))


### Build System

* **deps:** update dependencies ([#140](https://github.com/bucketeer-io/go-server-sdk/issues/140)) ([9a32609](https://github.com/bucketeer-io/go-server-sdk/commit/9a326094aa887f209d5069d9361da49c2f4b7d8f))

## [1.4.1](https://github.com/bucketeer-io/go-server-sdk/compare/v1.4.0...v1.4.1) (2024-06-13)


### Features

* new `Feature Flag` custom rule in the targeting settings ([#134](https://github.com/bucketeer-io/go-server-sdk/issues/134)) ([a71487d](https://github.com/bucketeer-io/go-server-sdk/commit/a71487da87ba8b08a6ace07479a4670e08d36741))

[See more information](https://docs.bucketeer.io/feature-flags/creating-feature-flags/targeting/#feature-flag)

## [1.4.0](https://github.com/bucketeer-io/go-server-sdk/compare/v1.3.6...v1.4.0) (2024-06-13)


### Features

* enable local evaluation ([#129](https://github.com/bucketeer-io/go-server-sdk/issues/129)) ([ea8092d](https://github.com/bucketeer-io/go-server-sdk/commit/ea8092d16990baff6bb181c718a49e3b7c070da6))


### Bug Fixes

* missing metrics in the cache processor ([#130](https://github.com/bucketeer-io/go-server-sdk/issues/130)) ([1ec9f89](https://github.com/bucketeer-io/go-server-sdk/commit/1ec9f89cac6c16dc64d7e42a32f13040e6e8c4b9))


### Miscellaneous

* add cache interface ([#119](https://github.com/bucketeer-io/go-server-sdk/issues/119)) ([ff593c8](https://github.com/bucketeer-io/go-server-sdk/commit/ff593c810c3ed9471250898ebff34b6520da7821))
* add codeowners file ([#116](https://github.com/bucketeer-io/go-server-sdk/issues/116)) ([8bb2693](https://github.com/bucketeer-io/go-server-sdk/commit/8bb26937c3edd93c9b0d4ab81e2a31eed7caa317))
* add feature flag and segment user cacher ([#121](https://github.com/bucketeer-io/go-server-sdk/issues/121)) ([df35501](https://github.com/bucketeer-io/go-server-sdk/commit/df35501500b55bd4d173cac89de3709231d0aa63))
* add feature flag cache processor ([#122](https://github.com/bucketeer-io/go-server-sdk/issues/122)) ([f65db04](https://github.com/bucketeer-io/go-server-sdk/commit/f65db04487cdaa9b22432899621ae8b52d410b25))
* add get feature flags and get segment users api ([#120](https://github.com/bucketeer-io/go-server-sdk/issues/120)) ([5d04d74](https://github.com/bucketeer-io/go-server-sdk/commit/5d04d7491bc304c33ef9332d57e56ecf184fe099))
* add metric events in the cache processors ([#127](https://github.com/bucketeer-io/go-server-sdk/issues/127)) ([6713101](https://github.com/bucketeer-io/go-server-sdk/commit/6713101a3754d64c3ac46d3ca73787f29804eb62))
* add new metrics events to handle unknown errors ([#115](https://github.com/bucketeer-io/go-server-sdk/issues/115)) ([19319b8](https://github.com/bucketeer-io/go-server-sdk/commit/19319b86e242180226b7f4cf175da98d65785d4a))
* add segment user cache processor ([#123](https://github.com/bucketeer-io/go-server-sdk/issues/123)) ([585afc3](https://github.com/bucketeer-io/go-server-sdk/commit/585afc3edaa639063d3b2f5c272b93b5add647b0))
* add SourceID and SDKVersion to the required request ([#118](https://github.com/bucketeer-io/go-server-sdk/issues/118)) ([fbf6e94](https://github.com/bucketeer-io/go-server-sdk/commit/fbf6e941e19ea2f4b14eaafc78898f1454e17591))
* export push event function so other processors report metric events ([#124](https://github.com/bucketeer-io/go-server-sdk/issues/124)) ([1ad1ecc](https://github.com/bucketeer-io/go-server-sdk/commit/1ad1ecc0b9b77b87bde87dcb47a543198d09c582))
* remove unused ctx paraemeter from event processor interface ([#125](https://github.com/bucketeer-io/go-server-sdk/issues/125)) ([b956faa](https://github.com/bucketeer-io/go-server-sdk/commit/b956faa22cfe4dc678316ae2d93054ac2312b484))


### Build System

* update dependencies ([#126](https://github.com/bucketeer-io/go-server-sdk/issues/126)) ([c6d0576](https://github.com/bucketeer-io/go-server-sdk/commit/c6d0576be341e22f84837ca21fe11c0a3ff69a9e))

## [1.3.6](https://github.com/bucketeer-io/go-server-sdk/compare/v1.3.5...v1.3.6) (2024-02-28)


### Miscellaneous

* update repository path ([#110](https://github.com/bucketeer-io/go-server-sdk/issues/110)) ([cae2883](https://github.com/bucketeer-io/go-server-sdk/commit/cae2883aeeee7a0e1f8f8bcf892371faa3e5a3e0))


### Build System

* update to go 1.20 and its dependencies ([#111](https://github.com/bucketeer-io/go-server-sdk/issues/111)) ([bf0af68](https://github.com/bucketeer-io/go-server-sdk/commit/bf0af681fb8d5ae7494ee25c2a6e41ff285231ca))

## [1.3.5](https://github.com/bucketeer-io/go-server-sdk/compare/v1.3.4...v1.3.5) (2023-06-23)


### Bug Fixes

* sdk version not being updated ([#99](https://github.com/bucketeer-io/go-server-sdk/issues/99)) ([9ee80fb](https://github.com/bucketeer-io/go-server-sdk/commit/9ee80fb6f65388b0a5454e400e80d12682064155))
* size metric event is reporting size incorrectly ([#102](https://github.com/bucketeer-io/go-server-sdk/issues/102)) ([61a9e39](https://github.com/bucketeer-io/go-server-sdk/commit/61a9e398d739a76e209985768382dbeeaf79ef5c))


### Miscellaneous

* support LatencySecond field ([#101](https://github.com/bucketeer-io/go-server-sdk/issues/101)) ([c38e0bb](https://github.com/bucketeer-io/go-server-sdk/commit/c38e0bb368759e00809c36ef32dc6a4f17e44b87))

## [1.3.4](https://github.com/bucketeer-io/go-server-sdk/compare/v1.3.3...v1.3.4) (2023-02-15)


### Miscellaneous

* request to grpc server ([#96](https://github.com/bucketeer-io/go-server-sdk/issues/96)) ([500a8b0](https://github.com/bucketeer-io/go-server-sdk/commit/500a8b0df74dfc08a339c25f3fb6733b977b0a23))
* support new metrics events ([#98](https://github.com/bucketeer-io/go-server-sdk/issues/98)) ([cc775c0](https://github.com/bucketeer-io/go-server-sdk/commit/cc775c0ee23d611e9f1ccd5b8a12591ae159a987))
