{ lib
, stdenv
, buildGo124Module
, subPackages ? null
, enableRaceDetector ? false
, enableTesting ? false
, parallelTesting ? false
, testGroup ? "all"
, jq
, solc
, solc-select
, clickhouse
, go-tools
, gotools
, golangci-lint
, delve
, gopls
, protoc-gen-go
, protobuf
, python3
}:
let inherit (lib) optional;
  overrideBuildGoModule = pkg: pkg.override { buildGoModule = buildGo124Module; };
in
buildGo124Module rec {
  name = "nil";
  pname = "nil";

  preBuild = ''
    make -j$NIX_BUILD_CORES generated rpcspec gen_rollup_contracts_bindings
    export HOME="$TMPDIR"
    mkdir -p ~/.gsolc-select/artifacts/solc-0.8.28
    ln -f -s ${solc}/bin/solc ~/.gsolc-select/artifacts/solc-0.8.28/solc-0.8.28

    if [ -z "$subPackages" ]; then
      case ${testGroup} in
        all|others) ;; # build everything
        *) subPackages=nil/internal/types;;
      esac
    fi
  '';

  tags = [ "assert" ];

  src = lib.sourceByRegex ./.. [
    "Makefile"
    "go.mod"
    "go.sum"
    "^nix(/tests-.*[.]txt)?$"
    "^nil(/.*)?$"
    "^smart-contracts(/.*)?$"
    "^uniswap(/.*)?$"
    "^rollup-bridge-contracts(/.*)?$"
  ];

  # to obtain run `nix build` with vendorHash = "";
  vendorHash = "sha256-7h5H7wPcz/SYNU7nWLNZzpnoGNb+LbD60jHTiNDE0Ig=";

  postInstall = ''
    mkdir -p $out/share/doc/nil
    cp openrpc.json $out/share/doc/nil

    mkdir -p $out/contracts/compiled
    cp -R nil/contracts/compiled $out/contracts/
  '';

  env = {
    CGO_ENABLED = if enableRaceDetector then 1 else 0;
  } // lib.optionalAttrs (subPackages != null) {
    subPackages = lib.concatStringsSep " " subPackages;
  };

  nativeBuildInputs = [
    jq
    solc
    solc-select
    clickhouse
    protobuf
    (overrideBuildGoModule gotools)
    (overrideBuildGoModule go-tools)
    (overrideBuildGoModule gopls)
    golangci-lint
    (overrideBuildGoModule delve)
    (overrideBuildGoModule protoc-gen-go)
    (python3.withPackages (ps: with ps; [
      safe-pysha3
    ]))
  ];

  packageName = "github.com/NilFoundation/nil";

  doCheck = enableTesting;
  checkFlags = [ "-tags assert,test,goexperiment.synctest" "-timeout 15m" ]
    ++ (if enableRaceDetector then [ "-race" ] else [ ]);

  preCheck = ''
    unset subPackages
    case ${testGroup} in
      all) ;; # build everything
      others)
        getGoDirs test | sed -e 's,^[.]/,,' | LC_ALL=C sort -u > all-tests.txt
        LC_ALL=C sort -u nix/tests-*.txt > all-groups.txt

        # run only tests that do not belong to a group
        subPackages=$(LC_ALL=C join -v 1 all-tests.txt all-groups.txt)
      ;;
      *) subPackages=$(cat nix/tests-${testGroup}.txt);; # run only required tests
    esac
    echo subPackages: $subPackages
  '';

  checkPhase = ''
    runHook preCheck
    # We do not set trimpath for tests, in case they reference test assets
    export GOFLAGS=''${GOFLAGS//-trimpath/}

    parallel=${if parallelTesting then "true" else "false"}
    if $parallel ; then
      buildGoDir test "$(getGoDirs test)"
    else
      for pkg in $(getGoDirs test); do
        buildGoDir test "$pkg"
      done
    fi

    runHook postCheck
  '';

  GOFLAGS = [ "-modcacherw" ];
  shellHook = ''
    eval "$configurePhase"
    export GOCACHE=/tmp/${vendorHash}/go-cache
    export GOMODCACHE=/tmp/${vendorHash}/go/mod/cache
    chmod -R u+w vendor
    mkdir -p ~/.solc-select/artifacts/solc-0.8.28
    ln -f -s ${solc}/bin/solc ~/.solc-select/artifacts/solc-0.8.28/solc-0.8.28
  '';
}
