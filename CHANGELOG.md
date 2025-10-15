# Changelog

## [0.2.7](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.6...v0.2.7) (2025-10-15)


### Bug Fixes

* karpenter panic when VM info is not yet populated ([91de0df](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/91de0df45620c6962dd33123307f1344541a9c76))

## [0.2.6](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.5...v0.2.6) (2025-10-14)


### Bug Fixes

* allow multiple Tags in ByTag selector ([e7ca6bc](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/e7ca6bc5eea92d064fb03c9e7902e51926c23197))

## [0.2.5](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.4...v0.2.5) (2025-10-14)


### Bug Fixes

* regenerate crd after userdata handling changes ([c58b0b6](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/c58b0b66fc39d402f518abf80ef9e64c25ee75c7))

## [0.2.4](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.3...v0.2.4) (2025-10-14)


### Bug Fixes

* fix garbage collection panics ([7a9657b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/7a9657bd4b3cc30e9f7710d82dfb1e4b70ff1658))

## [0.2.3](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.2...v0.2.3) (2025-10-03)


### Bug Fixes

* Add extraObjects to helm chart ([2125e1c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/2125e1cdde0569a8158d3f872dd34b5fae8e6086))

## [0.2.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.1...v0.2.2) (2025-10-03)


### Bug Fixes

* Set TAG env on correct action ([ce5cb26](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ce5cb26eb896da9da197cebbce5e7e308a62f84a))

## [0.2.1](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.0...v0.2.1) (2025-10-03)


### Bug Fixes

* Specify tag during ko build ([bc821d9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/bc821d917e6aa618c76b7475099901601af8072a))

## [0.2.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.1.0...v0.2.0) (2025-10-03)


### Features

* Refactor node bootstrap ([c087f5b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/c087f5b3c8df6ea6bdafd834077dc9e2f87f2f48))

## [0.1.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.5...v0.1.0) (2025-06-17)


### Features

* rework instsance types to include zone/region ([502cedd](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/502cedd3316944bbbcf8e24fdc20ee814813f3e1))


### Bug Fixes

* avoid nil reference ([3df5079](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/3df5079fb59d0a11f1c3e45c25ea4f38a795e583))

## [0.0.5](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.4...v0.0.5) (2025-06-16)


### Bug Fixes

* don't tag twice, fix image annotation ([f3517d6](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/f3517d65e42626767a1c2ac454c999020ed806c8))

## [0.0.4](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.3...v0.0.4) (2025-06-12)


### Bug Fixes

* Reuse finder instance ([98b9c2a](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/98b9c2a9ecc6d4f8f671bd690f515d8365ec3012))

## [0.0.3](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.2...v0.0.3) (2025-06-12)


### Miscellaneous Chores

* update README and minor cleanups ([18e1db7](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/18e1db7e52a945580e706377038c5945456fc45f))

## [0.0.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.2...v0.0.2) (2025-06-11)


### Features

* release 0.0.1 ([1660ada](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1660adaff997b63d0a17d15780061ed3b1dd1268))
* release pipeline ([aa963df](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/aa963df78727b39c3278b178809dec9bcc6c7037))
* set image tag ([d248378](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d2483782e5d708e04bf125a334d1986a9e45b7dd))
* test ([cad3a47](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/cad3a479601dd03540e0e3f14ee20e27402a03c0))
* test release-please-action ([4851a8b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/4851a8bc226490407567b0e0745bf2fe93053633))


### Bug Fixes

* adjust README ([ed97bab](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ed97bab875b6e2ddcaa2c5f78045eefa954a483d))
* chart path and build platform ([10c84e9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/10c84e978a4aa00989f00e747a2c4dbe3bf13288))
* don't try to change values.yaml ([e551673](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/e5516739e3cc00493dc08d187941f10fc9692d81))
* oci helm repo ([ebd2e99](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ebd2e99ff9db504a0f9fc5bde137c2b17ed6cc82))
* re-release ([60c40e2](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/60c40e26ea8b7f9f2e438c6a4de01d746fc0950e))
* release 0.0.1 ([936961c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/936961c9b8641b90fb2e842fd431bde5373b4832))
* release 0.0.1 ([50cd083](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/50cd083d3b908ab4d24c45196ea62da1daf1462b))
* release 0.0.1 ([2446fe0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/2446fe00ddb61e1e3575245699e7ebd605ee2bb7))
* release 0.0.2 ([8103063](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/810306313fe883a797072123f143cd023fbca8d1))
* release-please permissions ([eb9700d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/eb9700d9fbe73a037869e1cf32d2755375d89a9a))
* update chart version ([9ff036e](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9ff036ef92eea6505248929a4220c8017532c910))
* versioning ([1e90790](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1e907902a556e652ed1bb639c9d39fb4e33f6821))


### Miscellaneous Chores

* release 0.0.1 ([d39c354](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d39c35428d8cfe35f641151d0ec5d9a0fe4bd00c))
* **release:** release 0.0.1 ([1f82e60](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1f82e600653d3a93049b2cd9cdba27dd75c3d5c3))
* **release:** release 0.0.2 ([37b7135](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/37b7135a2c71c9e65730fae04bc878c012d1476a))
* **release:** release 0.0.2 ([9f33c36](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9f33c36abf6e4ed62fb5146300f83818b7a687a5))

## [0.0.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.2...v0.0.2) (2025-06-11)


### Features

* release 0.0.1 ([1660ada](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1660adaff997b63d0a17d15780061ed3b1dd1268))
* release pipeline ([aa963df](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/aa963df78727b39c3278b178809dec9bcc6c7037))
* set image tag ([d248378](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d2483782e5d708e04bf125a334d1986a9e45b7dd))
* test ([cad3a47](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/cad3a479601dd03540e0e3f14ee20e27402a03c0))
* test release-please-action ([4851a8b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/4851a8bc226490407567b0e0745bf2fe93053633))


### Bug Fixes

* adjust README ([ed97bab](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ed97bab875b6e2ddcaa2c5f78045eefa954a483d))
* chart path and build platform ([10c84e9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/10c84e978a4aa00989f00e747a2c4dbe3bf13288))
* don't try to change values.yaml ([e551673](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/e5516739e3cc00493dc08d187941f10fc9692d81))
* oci helm repo ([ebd2e99](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ebd2e99ff9db504a0f9fc5bde137c2b17ed6cc82))
* release 0.0.1 ([936961c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/936961c9b8641b90fb2e842fd431bde5373b4832))
* release 0.0.1 ([50cd083](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/50cd083d3b908ab4d24c45196ea62da1daf1462b))
* release 0.0.1 ([2446fe0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/2446fe00ddb61e1e3575245699e7ebd605ee2bb7))
* release 0.0.2 ([8103063](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/810306313fe883a797072123f143cd023fbca8d1))
* release-please permissions ([eb9700d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/eb9700d9fbe73a037869e1cf32d2755375d89a9a))
* update chart version ([9ff036e](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9ff036ef92eea6505248929a4220c8017532c910))
* versioning ([1e90790](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1e907902a556e652ed1bb639c9d39fb4e33f6821))


### Miscellaneous Chores

* release 0.0.1 ([d39c354](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d39c35428d8cfe35f641151d0ec5d9a0fe4bd00c))
* **release:** release 0.0.1 ([1f82e60](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1f82e600653d3a93049b2cd9cdba27dd75c3d5c3))
* **release:** release 0.0.2 ([9f33c36](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9f33c36abf6e4ed62fb5146300f83818b7a687a5))

## [0.0.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.0.2...v0.0.2) (2025-06-11)


### Features

* release 0.0.1 ([1660ada](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1660adaff997b63d0a17d15780061ed3b1dd1268))
* release pipeline ([aa963df](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/aa963df78727b39c3278b178809dec9bcc6c7037))
* set image tag ([d248378](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d2483782e5d708e04bf125a334d1986a9e45b7dd))
* test ([cad3a47](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/cad3a479601dd03540e0e3f14ee20e27402a03c0))
* test release-please-action ([4851a8b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/4851a8bc226490407567b0e0745bf2fe93053633))


### Bug Fixes

* chart path and build platform ([10c84e9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/10c84e978a4aa00989f00e747a2c4dbe3bf13288))
* don't try to change values.yaml ([e551673](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/e5516739e3cc00493dc08d187941f10fc9692d81))
* oci helm repo ([ebd2e99](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ebd2e99ff9db504a0f9fc5bde137c2b17ed6cc82))
* release 0.0.1 ([936961c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/936961c9b8641b90fb2e842fd431bde5373b4832))
* release 0.0.1 ([50cd083](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/50cd083d3b908ab4d24c45196ea62da1daf1462b))
* release 0.0.1 ([2446fe0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/2446fe00ddb61e1e3575245699e7ebd605ee2bb7))
* release 0.0.2 ([8103063](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/810306313fe883a797072123f143cd023fbca8d1))
* release-please permissions ([eb9700d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/eb9700d9fbe73a037869e1cf32d2755375d89a9a))
* update chart version ([9ff036e](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9ff036ef92eea6505248929a4220c8017532c910))
* versioning ([1e90790](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1e907902a556e652ed1bb639c9d39fb4e33f6821))


### Miscellaneous Chores

* release 0.0.1 ([d39c354](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d39c35428d8cfe35f641151d0ec5d9a0fe4bd00c))
* **release:** release 0.0.1 ([1f82e60](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1f82e600653d3a93049b2cd9cdba27dd75c3d5c3))

## 0.0.1 (2025-06-11)


### Features

* release 0.0.1 ([1660ada](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/1660adaff997b63d0a17d15780061ed3b1dd1268))
* release pipeline ([aa963df](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/aa963df78727b39c3278b178809dec9bcc6c7037))
* test release-please-action ([4851a8b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/4851a8bc226490407567b0e0745bf2fe93053633))


### Bug Fixes

* chart path and build platform ([10c84e9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/10c84e978a4aa00989f00e747a2c4dbe3bf13288))
* don't try to change values.yaml ([e551673](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/e5516739e3cc00493dc08d187941f10fc9692d81))
* release 0.0.1 ([936961c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/936961c9b8641b90fb2e842fd431bde5373b4832))
* release 0.0.1 ([50cd083](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/50cd083d3b908ab4d24c45196ea62da1daf1462b))
* release 0.0.1 ([2446fe0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/2446fe00ddb61e1e3575245699e7ebd605ee2bb7))
* release-please permissions ([eb9700d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/eb9700d9fbe73a037869e1cf32d2755375d89a9a))


### Miscellaneous Chores

* release 0.0.1 ([d39c354](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d39c35428d8cfe35f641151d0ec5d9a0fe4bd00c))
