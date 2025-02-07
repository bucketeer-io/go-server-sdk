# Changelog

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
