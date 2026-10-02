#!/bin/bash

# Exercise the actual installer without network access or architecture-specific binaries.
set -euo pipefail

installer="$(cd "$(dirname "$0")/../install" && pwd)/install-mac.sh"
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT

mkdir -p "$test_root/bin" "$test_root/fixtures"
cat > "$test_root/bin/uname" <<'EOF'
#!/bin/bash
echo "$TEST_CPU"
EOF
cat > "$test_root/bin/sysctl" <<'EOF'
#!/bin/bash
[[ "$*" == "-in sysctl.proc_translated" ]] || exit 99
[[ "$TEST_TRANSLATED" != unavailable ]] || exit 1
echo "$TEST_TRANSLATED"
EOF
cat > "$test_root/bin/curl" <<'EOF'
#!/bin/bash
set -euo pipefail
[[ $# == 3 && "$1" == -fsSLo ]] || exit 99
destination=$2
url=$3
echo "$url" >> "$TEST_DOWNLOADS"
case "$url" in
    "$TEST_MIRROR/download/$TEST_VERSION/porter-darwin-amd64") asset=amd64 ;;
    "$TEST_MIRROR/download/$TEST_VERSION/porter-darwin-arm64") asset=arm64 ;;
    "$TEST_MIRROR/download/$TEST_VERSION/porter-linux-amd64") asset=runtime ;;
    *) echo "Unexpected URL: $url" >&2; exit 99 ;;
esac
if [[ "$TEST_FAIL_DOWNLOAD" == "$asset" ]]; then
    echo partial > "$destination"
    echo 'curl: requested asset unavailable' >&2
    exit 22
fi
cp "$TEST_FIXTURES/$asset" "$destination"
EOF
cat > "$test_root/bin/mv" <<'EOF'
#!/bin/bash
set -euo pipefail
for destination; do :; done
if [[ "$TEST_FAIL_RENAME" == runtime && "$destination" == "$PORTER_HOME/runtimes/porter-runtime" ]] ||
   [[ "$TEST_FAIL_RENAME" == cli && "$destination" == "$PORTER_HOME/porter" ]]; then
    echo 'mv: simulated rename failure' >&2
    exit 73
fi
/bin/mv "$@"
EOF
cat > "$test_root/bin/chmod" <<'EOF'
#!/bin/bash
[[ "$TEST_FAIL_CHMOD" == 0 ]] || exit 74
/bin/chmod "$@"
EOF

for arch in amd64 arm64; do
    echo '#!/bin/bash' > "$test_root/fixtures/$arch"
    echo "arch=$arch" >> "$test_root/fixtures/$arch"
    cat >> "$test_root/fixtures/$arch" <<'EOF'
set -euo pipefail
if [[ $# == 1 && "$1" == version ]]; then
    if [[ "$TEST_CHECK_PRESERVED" == 1 ]]; then
        cmp "$PORTER_HOME/porter" "$TEST_OLD_CLI"
        cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_OLD_RUNTIME"
    fi
    [[ "$TEST_FAIL_VERSION" == 0 ]] || exit 42
    echo "porter $TEST_VERSION darwin/$arch"
elif [[ $# == 5 && "$1 $2 $3 $4" == 'mixin install exec --version' && "$5" == "$TEST_VERSION" ]]; then
    [[ "$0" == "$PORTER_HOME/porter" ]] || exit 98
    cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_FIXTURES/runtime"
    [[ "$TEST_FAIL_MIXIN" == 0 ]] || exit 43
    mkdir -p "$PORTER_HOME/mixins/exec"
    printf '#!/bin/bash\necho "exec %s darwin/%s"\n' "$TEST_VERSION" "$arch" > "$PORTER_HOME/mixins/exec/exec"
    /bin/chmod +x "$PORTER_HOME/mixins/exec/exec"
else
    echo "Unexpected CLI arguments: $*" >&2
    exit 99
fi
EOF
done
printf '#!/bin/bash\necho "runtime linux/amd64"\n' > "$test_root/fixtures/runtime"
/bin/chmod +x "$test_root/bin/"* "$test_root/fixtures/"*

setup() {
    export PATH="$test_root/bin:/usr/bin:/bin:/usr/sbin:/sbin"
    export HOME="$test_root/$1/home"
    export PORTER_HOME="$HOME/.porter"
    unset PORTER_VERSION PORTER_MIRROR
    export TEST_CPU=x86_64 TEST_TRANSLATED=0
    export TEST_VERSION=latest TEST_MIRROR=https://cdn.porter.sh
    export TEST_FAIL_DOWNLOAD=none TEST_FAIL_VERSION=0 TEST_FAIL_MIXIN=0
    export TEST_FAIL_RENAME=none TEST_FAIL_CHMOD=0 TEST_CHECK_PRESERVED=0
    export TEST_FIXTURES="$test_root/fixtures"
    export TEST_DOWNLOADS="$test_root/$1/downloads"
    export TEST_OLD_CLI="$test_root/$1/old-cli"
    export TEST_OLD_RUNTIME="$test_root/$1/old-runtime"
    output="$test_root/$1/output"
    mkdir -p "$HOME"
    : > "$TEST_DOWNLOADS"
}

run_installer() {
    status=0
    /bin/bash "$installer" > "$output" 2>&1 || status=$?
    cat "$output"
}

assert_equal() {
    if [[ "$1" != "$2" ]]; then
        echo "Expected '$2', got '$1'" >&2
        exit 1
    fi
}

assert_clean() {
    if [[ -d "$PORTER_HOME" ]]; then
        # Only installed assets should remain, regardless of how staging is named.
        assert_equal "$(find "$PORTER_HOME" -type f ! -path "$PORTER_HOME/porter" ! -path "$PORTER_HOME/runtimes/porter-runtime" ! -path "$PORTER_HOME/mixins/exec/exec")" ''
    fi
}

assert_failed() {
    assert_equal "$status" "$1"
    if grep -Eq 'Installed |Installation complete|export PATH=|Add porter to your path' "$output"; then
        echo 'Installer reported success after failure' >&2
        exit 1
    fi
    assert_clean
}

assert_installed() {
    assert_equal "$status" 0
    [[ -x "$PORTER_HOME/porter" && -x "$PORTER_HOME/runtimes/porter-runtime" ]]
    cmp "$PORTER_HOME/porter" "$TEST_FIXTURES/$1"
    cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_FIXTURES/runtime"
    TEST_CHECK_PRESERVED=0
    assert_equal "$("$PORTER_HOME/porter" version)" "porter $TEST_VERSION darwin/$1"
    assert_equal "$("$PORTER_HOME/mixins/exec/exec" version)" "exec $TEST_VERSION darwin/$1"
    assert_equal "$("$PORTER_HOME/runtimes/porter-runtime")" 'runtime linux/amd64'
    grep -Fq "porter $TEST_VERSION darwin/$1" "$output"
    grep -Fq 'Installation complete.' "$output"
    assert_clean
}

seed_installation() {
    mkdir -p "$PORTER_HOME/runtimes"
    printf '#!/bin/bash\necho old-cli\n' > "$TEST_OLD_CLI"
    printf '#!/bin/bash\necho old-runtime\n' > "$TEST_OLD_RUNTIME"
    cp "$TEST_OLD_CLI" "$PORTER_HOME/porter"
    cp "$TEST_OLD_RUNTIME" "$PORTER_HOME/runtimes/porter-runtime"
    /bin/chmod +x "$PORTER_HOME/porter" "$PORTER_HOME/runtimes/porter-runtime"
}

assert_preserved() {
    cmp "$PORTER_HOME/porter" "$TEST_OLD_CLI"
    cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_OLD_RUNTIME"
    assert_equal "$("$PORTER_HOME/porter" version)" old-cli
    assert_equal "$("$PORTER_HOME/runtimes/porter-runtime")" old-runtime
}

test_architecture() {
    TEST_CPU=$1 TEST_TRANSLATED=$2
    run_installer
    assert_installed "$3"
}

test_unsupported() {
    TEST_CPU=ppc64
    run_installer
    assert_failed 1
    grep -Fq ppc64 "$output"
    [[ ! -e "$PORTER_HOME" && ! -s "$TEST_DOWNLOADS" ]]
}

test_missing_arm64() {
    TEST_CPU=arm64 TEST_FAIL_DOWNLOAD=arm64
    export PORTER_VERSION=v0.23.0-beta.1
    TEST_VERSION=$PORTER_VERSION
    seed_installation
    run_installer
    assert_failed 22
    assert_preserved
    grep -Fq "$TEST_VERSION" "$output"
    grep -Fq arm64 "$output"
    grep -Fq 'https://cdn.porter.sh/download/v0.23.0-beta.1/porter-darwin-arm64' "$output"
    if grep -Fq porter-darwin-amd64 "$TEST_DOWNLOADS"; then exit 1; fi
}

test_download_failure() {
    TEST_FAIL_DOWNLOAD=$1
    seed_installation
    run_installer
    assert_failed 22
    assert_preserved
}

test_version_failure() {
    TEST_FAIL_VERSION=1
    seed_installation
    run_installer
    assert_failed 42
    assert_preserved
}

test_chmod_failure() {
    TEST_FAIL_CHMOD=1
    seed_installation
    run_installer
    assert_failed 74
    assert_preserved
}

test_rename_failure() {
    TEST_FAIL_RENAME=$1
    seed_installation
    run_installer
    assert_failed 73
    cmp "$PORTER_HOME/porter" "$TEST_OLD_CLI"
    if [[ "$1" == runtime ]]; then
        cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_OLD_RUNTIME"
    else
        cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_FIXTURES/runtime"
    fi
}

test_mixin_failure() {
    TEST_FAIL_MIXIN=1
    run_installer
    assert_failed 43
    cmp "$PORTER_HOME/porter" "$TEST_FIXTURES/amd64"
    cmp "$PORTER_HOME/runtimes/porter-runtime" "$TEST_FIXTURES/runtime"
    TEST_FAIL_MIXIN=0
    run_installer
    assert_installed amd64
}

test_custom_settings() {
    export PORTER_HOME="$HOME/custom porter" PORTER_VERSION=v1.6.1
    export PORTER_MIRROR='https://mirror.example/porter assets'
    TEST_VERSION=$PORTER_VERSION TEST_MIRROR=$PORTER_MIRROR
    run_installer
    assert_installed amd64
    # The printed PATH command must work when pasted into the user's shell.
    path_instruction=$(sed -n '/^export PATH=/p' "$output")
    [[ -n "$path_instruction" ]]
    eval "$path_instruction"
    assert_equal "$(command -v porter)" "$PORTER_HOME/porter"
}

test_default_home() {
    HOME="$HOME/with spaces"
    unset PORTER_HOME
    run_installer
    export PORTER_HOME="$HOME/.porter"
    assert_installed amd64
}

test_repeated_install() {
    seed_installation
    TEST_CHECK_PRESERVED=1
    run_installer
    assert_installed amd64
    run_installer
    assert_installed amd64
}

passed=0
failed=0
run_test() {
    local name=$1
    shift
    set +e
    (
        set -e
        setup "$name"
        "$@"
    ) > "$test_root/$name.log" 2>&1
    local result=$?
    set -e
    if [[ "$result" == 0 ]]; then
        echo "PASS $name"
        passed=$((passed + 1))
    else
        echo "FAIL $name"
        cat "$test_root/$name.log"
        failed=$((failed + 1))
    fi
}

run_test intel test_architecture x86_64 0 amd64
run_test intel_unavailable_sysctl test_architecture x86_64 unavailable amd64
run_test intel_empty_sysctl test_architecture x86_64 '' amd64
run_test intel_other_sysctl test_architecture x86_64 2 amd64
run_test apple_silicon test_architecture arm64 unavailable arm64
run_test rosetta test_architecture x86_64 1 arm64
run_test unsupported test_unsupported
run_test missing_arm64 test_missing_arm64
run_test cli_download_failure test_download_failure amd64
run_test runtime_download_failure test_download_failure runtime
run_test version_failure test_version_failure
run_test chmod_failure test_chmod_failure
run_test runtime_rename_failure test_rename_failure runtime
run_test cli_rename_failure test_rename_failure cli
run_test mixin_failure_and_retry test_mixin_failure
run_test custom_settings test_custom_settings
run_test default_home_spaces test_default_home
run_test repeated_install test_repeated_install

echo "$passed passed, $failed failed"
[[ "$failed" == 0 ]]
