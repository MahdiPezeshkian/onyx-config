# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- `All` and `GetMap` return deep copies; `GetStringSlice` copies `[]string` values.
- Documented BindEnv resolution, concurrency, and flat XML limitations.

### Added
- `LookupString`, `LookupInt`, `LookupBool` with `ErrNotFound` / `ErrTypeMismatch`.
- Nil-safe `Merge`.


## [0.1.0] - 2026-11-11

### Added

- Initial extraction of the configuration loader from `Onyx.Common/configmaster`.
- Multi-format loading (YAML, TOML, JSON, XML, ENV), defaults, merging, typed getters, and env binding.

[Unreleased]: https://github.com/MahdiPezeshkian/onyx-config/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/MahdiPezeshkian/onyx-config/releases/tag/v0.1.0
