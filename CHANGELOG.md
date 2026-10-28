# Changelog

## [3.2.5](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.2.4...v3.2.5) (2026-10-28)


### Bug Fixes

* **deps:** update github.com/awslabs/operatorpkg digest to 4e7ba6e ([#124](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/124)) ([0e0e699](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/0e0e699c42e09911a66e0dcf26c97ad2efd88169))
* **deps:** update kubernetes dependencies to v0.37.2 ([#132](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/132)) ([9f235a8](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/9f235a898c49cff7fef4b8413eb585451cda44aa))
* **deps:** update module github.com/vmware/govmomi to v0.56.1 ([#129](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/129)) ([0f924a3](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/0f924a3f00bac9f496e2ad42f175139e2c379ccd))
* **deps:** update module github.com/vmware/govmomi to v0.56.2 ([#130](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/130)) ([3461c8a](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/3461c8a123e77eed7d8765196a102ecf54aa12d5))
* **deps:** update module sigs.k8s.io/karpenter to v1.15.1 ([#127](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/127)) ([19fbdb0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/19fbdb0fff2512a9378ed0c6c03a84791fccdad6))


### Miscellaneous Chores

* **deps:** update docker/login-action action to v4.7.0 ([#128](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/128)) ([fa75472](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/fa75472555fa1c3742e73cf930cd36d1491fb395))

## [3.2.4](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.2.3...v3.2.4) (2026-10-17)


### Bug Fixes

* **deps:** update github.com/awslabs/operatorpkg digest to 855c7d0 ([16c1c0e](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/16c1c0e0571d7a4c6581527cf96d39d59971e94f))
* **deps:** update kubernetes dependencies to v0.37.1 ([8a8b1f4](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/8a8b1f4b616908014a8506fd4ba89b6f2f23ba09))
* **deps:** update module github.com/coreos/ignition/v2 to v2.28.0 ([#119](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/119)) ([070d421](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/070d421bcb405c5c7161e27943f3a50f0c20eb47))
* **deps:** update module github.com/stretchr/testify to v1.12.1 ([#120](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/120)) ([c787d61](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/c787d61b913b1aca30473b5ded5a1bf555318e3d))
* **deps:** update module github.com/vmware/govmomi to v0.56.0 ([#121](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/121)) ([c4b0a87](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/c4b0a870f9fe70a6d40bb94b73257e5d9404bb79))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.25.2 ([#122](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/122)) ([5ce2f7b](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/5ce2f7b45bf3b8c73528af9c47b2c1d5a8fe92ee))
* **deps:** update module sigs.k8s.io/karpenter to v1.15.0 ([555bf58](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/555bf5864ed15276300296950dd7436791da585e))
* GenerateNodeClaimName returning empty name ([8dffe98](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/8dffe98a375274cf8b7e0aad21f106021f0b0944))


### Miscellaneous Chores

* configure Renovate ([b038cd7](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/b038cd7b1de105454a9a399eb9acc1553207c032))
* **deps:** update azure/setup-helm action to v4.3.1 ([6bcba8a](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/6bcba8af25d1fbae00e5d82ac8b5d73b7595bc1b))
* **deps:** update github actions dependencies ([972a8d0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/972a8d00765381095a7fcc81bd6db05e310af67b))
* **deps:** update github actions dependencies ([bea47f0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/bea47f044c9e56e79fe40f14df001249b87c1598))
* **fix:** build against karpenter 1.15 ([78f713e](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/78f713ef695226ad577dd4c11010b96ccf3dd92a))
* regenerate mocks ([a3bc936](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/a3bc936e154732267b455efcba4429311e59831f))

## [3.2.3](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.2.2...v3.2.3) (2026-10-16)


### Bug Fixes

* Sync nodepool crd ([8a2daaf](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/8a2daaf41bcc017a41d32787211728d081490db1))

## [3.2.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.2.1...v3.2.2) (2026-10-16)


### Bug Fixes

* Add zone, region settings as helmchart value ([879f0c4](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/879f0c41534bafa173a79ed846d2923de6ce6738))

## [3.2.1](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.2.0...v3.2.1) (2026-10-16)


### Bug Fixes

* drop CEL validation rules ([0b01e4a](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/0b01e4a400df569ab5a753ef9544b0a38c0e1c70))

## [3.2.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.1.0...v3.2.0) (2026-10-16)


### Features

* KwoK instance types ([#93](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/93)) ([6699708](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/66997089f544b7756be0a9a68d808163874970de))
* vcsim simulator and finder_test ([#83](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/83)) ([aed25b0](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/aed25b022fb85a7a5228c7e7ce14a71b16debf8d))


### Bug Fixes

* change cloned_from annotation format ([6815d32](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/6815d32aa52d064b7d1789730771c06fbcfa04e1))
* finder, instance and session e2e vcsim tests ([99a0239](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/99a0239e3ec65eb764b14f75094e09e826064ce5))
* Karpenter capacity type OnDemand replaced with on-demand ([b46da1c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/b46da1c3c3d68bee36e0ecbd42ad055da40b8ba5))
* set capacity type to offerings ([#92](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/92)) ([61daaf9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/61daaf9c2361f2edd9e3f03e579734e64c861773))
* validate kubeReserved values and rename KubeletConfiguration ([d67279c](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d67279c49e6c2dde0287afa3d003cadf62b1424b))

## [3.1.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.0.2...v3.1.0) (2026-08-27)


### Features

* Bump dependencies, add undecorated provider ([d7e2d55](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d7e2d55f0718c0971a39786c97e77aa9a4bee48c))


### Bug Fixes

* .gitignore ([be34181](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/be3418141d28188e757d637c6e2b271fbceb1802))
* implement InstanceTypeToMegabytes utility and update memory calculations in instance provider ([12afe0f](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/12afe0fb9e2b6f70fdcceb69f4ac2298aeaf77a8))

## [3.0.2](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.0.1...v3.0.2) (2026-08-12)


### Bug Fixes

* fix error expression. ([#74](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/74)) ([02d0589](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/02d05894584ecd8546c22f899646ed094da5dd7a))

## [3.0.1](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v3.0.0...v3.0.1) (2026-08-10)


### Bug Fixes

* re-authenticate vSphere session on expiry instead of failing with 401 ([bece561](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/bece56101478ec6ca42b810b07c4330435b0586f))

## [3.0.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v2.0.1...v3.0.0) (2026-07-02)


### ⚠ BREAKING CHANGES

* Add vsphere and kubeapi as chart values

### Code Refactoring

* Add vsphere and kubeapi as chart values ([b52f798](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/b52f798817075970fa4f47c85a5c8f3f86dc7f15))

## [2.0.1](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v2.0.0...v2.0.1) (2026-06-11)


### Bug Fixes

* set disk.EnableUUID=TRUE ([ce8da36](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/ce8da361cfe231d29b4333cce4d6b7c7e84990f9))

## [2.0.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v1.0.0...v2.0.0) (2026-06-01)


### ⚠ BREAKING CHANGES

* userdata hadnling distros and ignition

### Code Refactoring

* userdata hadnling distros and ignition ([69f85d4](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/69f85d4aa2d2670d678ad5fa5ef09cc3ef769151))

## [1.0.0](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.13...v1.0.0) (2026-04-07)


### ⚠ BREAKING CHANGES

* Multi DC machine spread is almost never used use case, move DC configuration into controller options.

### Code Refactoring

* Move DC configuraion from nodeclass to controller ([f07675d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/f07675de4844c0bebb22cad221846bd1fcf366c1))

## [0.2.13](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.12...v0.2.13) (2026-03-04)


### Miscellaneous Chores

* bump go to 1.24.13 ([cb2ce2d](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/cb2ce2d77dfd02394b03a72655abf8c4dbb938a0))

## [0.2.12](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.11...v0.2.12) (2026-01-16)


### Bug Fixes

* Cover more network type in byTag lookup ([3c609e9](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/3c609e90ab81b6142823b2aa7c260a04904d1b00))

## [0.2.11](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.10...v0.2.11) (2026-01-16)


### Bug Fixes

* Network by Tag lookup ([cd45ab3](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/cd45ab312b64029cd0ae9585204d3db1b9a5d6a0))

## [0.2.10](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.9...v0.2.10) (2025-12-01)


### Bug Fixes

* README updates and code fixes ([4293c39](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/4293c394b56a2a14fb3ac0728d69937bfcd865f3))

## [0.2.9](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.8...v0.2.9) (2025-10-15)


### Bug Fixes

* RKE2 config node-taint typo ([b7bfcfd](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/b7bfcfdda17c581f6725c3b0411d8c0901e3a415))

## [0.2.8](https://github.com/AbsaOSS/karpenter-provider-vsphere/compare/v0.2.7...v0.2.8) (2025-10-15)


### Bug Fixes

* Race condition in a create ([#42](https://github.com/AbsaOSS/karpenter-provider-vsphere/issues/42)) ([d21ab40](https://github.com/AbsaOSS/karpenter-provider-vsphere/commit/d21ab408a9c135f6fd9af4f0ffd6f58874211158))

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
