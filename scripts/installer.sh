#!/usr/bin/env bash

# Copyright The Helm Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The install script is based off of the MIT-licensed script from glide,
# the package manager for Go: https://github.com/Masterminds/glide.sh/blob/master/get

: ${BINARY_NAME:="xt"}
: ${USE_SUDO:="true"}
: ${DEBUG:="false"}
: ${VERIFY_CHECKSUM:="true"}
: ${VERIFY_SIGNATURES:="false"}
: ${XT_INSTALL_DIR:="/usr/local/bin"}
: ${GPG_PUBRING:="pubring.kbx"}

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"
HAS_OPENSSL="$(type "openssl" &> /dev/null && echo true || echo false)"
HAS_GPG="$(type "gpg" &> /dev/null && echo true || echo false)"

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
  esac
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  local CMD="$*"

  if [ $EUID -ne 0 -a $USE_SUDO = "true" ]; then
    CMD="sudo $CMD"
  fi

  $CMD
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="darwin-amd64\nlinux-386\nlinux-amd64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/adamkobi/xt"
    exit 1
  fi

  if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
    echo "Either curl or wget is required"
    exit 1
  fi

  if [ "${VERIFY_CHECKSUM}" == "true" ] && [ "${HAS_OPENSSL}" != "true" ]; then
    echo "In order to verify checksum, openssl must first be installed."
    echo "Please install openssl or set VERIFY_CHECKSUM=false in your environment."
    exit 1
  fi

  if [ "${VERIFY_SIGNATURES}" == "true" ]; then
    if [ "${HAS_GPG}" != "true" ]; then
      echo "In order to verify signatures, gpg must first be installed."
      echo "Please install gpg or set VERIFY_SIGNATURES=false in your environment."
      exit 1
    fi
    if [ "${OS}" != "linux" ]; then
      echo "Signature verification is currently only supported on Linux."
      echo "Please set VERIFY_SIGNATURES=false or verify the signatures manually."
      exit 1
    fi
  fi
}

# checkDesiredVersion checks if the desired version is available.
checkDesiredVersion() {
  if [ "x$DESIRED_VERSION" == "x" ]; then
    # Get tag from release URL
    local latest_release_url="https://github.com/adamkobi/xt/releases"
    if [ "${HAS_CURL}" == "true" ]; then
      TAG=$(curl -Ls $latest_release_url | grep 'href="/adamkobi/xt/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | grep -v no-underline | head -n 1 | cut -d '"' -f 2 | awk '{n=split($NF,a,"/");print a[n]}' | awk 'a !~ $0{print}; {a=$0}')
    elif [ "${HAS_WGET}" == "true" ]; then
      TAG=$(wget $latest_release_url -O - 2>&1 | grep 'href="/adamkobi/xt/releases/tag/v[0-9]*.[0-9]*.[0-9]*\"' | grep -v no-underline | head -n 1 | cut -d '"' -f 2 | awk '{n=split($NF,a,"/");print a[n]}' | awk 'a !~ $0{print}; {a=$0}')
    fi
  else
    TAG=$DESIRED_VERSION
  fi
}

# checkXtInstalledVersion checks which version of xt is installed and
# if it needs to be changed.
checkXtInstalledVersion() {
  if [[ -f "${XT_INSTALL_DIR}/${BINARY_NAME}" ]]; then
    local version=$("${XT_INSTALL_DIR}/${BINARY_NAME}" version | head -n 1 | cut -d' ' -f3)
    if [[ "$version" == "$TAG" ]]; then
      echo "Xt ${version} is already ${DESIRED_VERSION:-latest}"
      return 0
    else
      echo "Xt ${TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package and also the checksum
# for that binary.
downloadFile() {
  XT_DIST="xt-$TAG-$OS-$ARCH.tar.gz"
  DOWNLOAD_URL="https://github.com/adamkobi/xt/releases/$XT_DIST"
  CHECKSUM_URL="$DOWNLOAD_URL.sha256"
  XT_TMP_ROOT="$(mktemp -dt xt-installer-XXXXXX)"
  XT_TMP_FILE="$XT_TMP_ROOT/$XT_DIST"
  XT_SUM_FILE="$XT_TMP_ROOT/$XT_DIST.sha256"
  echo "Downloading $DOWNLOAD_URL"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "$CHECKSUM_URL" -o "$XT_SUM_FILE"
    curl -SsL "$DOWNLOAD_URL" -o "$XT_TMP_FILE"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "$XT_SUM_FILE" "$CHECKSUM_URL"
    wget -q -O "$XT_TMP_FILE" "$DOWNLOAD_URL"
  fi
}

# verifyFile verifies the SHA256 checksum of the binary package
# and the GPG signatures for both the package and checksum file
# (depending on settings in environment).
verifyFile() {
  if [ "${VERIFY_CHECKSUM}" == "true" ]; then
    verifyChecksum
  fi
  if [ "${VERIFY_SIGNATURES}" == "true" ]; then
    verifySignatures
  fi
}

# installFile installs the Xt binary.
installFile() {
  XT_TMP="$XT_TMP_ROOT/$BINARY_NAME"
  mkdir -p "$XT_TMP"
  tar xf "$XT_TMP_FILE" -C "$XT_TMP"
  XT_TMP_BIN="$XT_TMP/$OS-$ARCH/xt"
  echo "Preparing to install $BINARY_NAME into ${XT_INSTALL_DIR}"
  runAsRoot cp "$XT_TMP_BIN" "$XT_INSTALL_DIR/$BINARY_NAME"
  echo "$BINARY_NAME installed into $XT_INSTALL_DIR/$BINARY_NAME"
}

# verifyChecksum verifies the SHA256 checksum of the binary package.
verifyChecksum() {
  printf "Verifying checksum... "
  local sum=$(openssl sha1 -sha256 ${XT_TMP_FILE} | awk '{print $2}')
  local expected_sum=$(cat ${XT_SUM_FILE})
  echo ${expected_sum}
  echo ${sum}
  if [ "$sum" != "$expected_sum" ]; then
    echo "SHA sum of ${XT_TMP_FILE} does not match. Aborting."
    exit 1
  fi
  echo "Done."
}

# verifySignatures obtains the latest KEYS file from GitHub master branch
# as well as the signature .asc files from the specific GitHub release,
# then verifies that the release artifacts were signed by a maintainer's key.
verifySignatures() {
  printf "Verifying signatures... "
  local keys_filename="KEYS"
  local github_keys_url="https://raw.githubusercontent.com/adamkobi/xt/master/${keys_filename}"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "${github_keys_url}" -o "${XT_TMP_ROOT}/${keys_filename}"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "${XT_TMP_ROOT}/${keys_filename}" "${github_keys_url}"
  fi
  local gpg_keyring="${XT_TMP_ROOT}/keyring.gpg"
  local gpg_homedir="${XT_TMP_ROOT}/gnupg"
  mkdir -p -m 0700 "${gpg_homedir}"
  local gpg_stderr_device="/dev/null"
  if [ "${DEBUG}" == "true" ]; then
    gpg_stderr_device="/dev/stderr"
  fi
  gpg --batch --quiet --homedir="${gpg_homedir}" --import "${XT_TMP_ROOT}/${keys_filename}" 2> "${gpg_stderr_device}"
  gpg --batch --no-default-keyring --keyring "${gpg_homedir}/${GPG_PUBRING}" --export > "${gpg_keyring}"
  local github_release_url="https://github.com/adamkobi/xt/releases/download/${TAG}"
  if [ "${HAS_CURL}" == "true" ]; then
    curl -SsL "${github_release_url}/xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" -o "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc"
    curl -SsL "${github_release_url}/xt-${TAG}-${OS}-${ARCH}.tar.gz.asc" -o "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.asc"
  elif [ "${HAS_WGET}" == "true" ]; then
    wget -q -O "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" "${github_release_url}/xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc"
    wget -q -O "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.asc" "${github_release_url}/xt-${TAG}-${OS}-${ARCH}.tar.gz.asc"
  fi
  local error_text="If you think this might be a potential security issue,"
  error_text="${error_text}\nplease see here: https://github.com/adamkobi/community/blob/master/SECURITY.md"
  local num_goodlines_sha=$(gpg --verify --keyring="${gpg_keyring}" --status-fd=1 "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256.asc" 2> "${gpg_stderr_device}" | grep -c -E '^\[GNUPG:\] (GOODSIG|VALIDSIG)')
  if [[ ${num_goodlines_sha} -lt 2 ]]; then
    echo "Unable to verify the signature of xt-${TAG}-${OS}-${ARCH}.tar.gz.sha256!"
    echo -e "${error_text}"
    exit 1
  fi
  local num_goodlines_tar=$(gpg --verify --keyring="${gpg_keyring}" --status-fd=1 "${XT_TMP_ROOT}/xt-${TAG}-${OS}-${ARCH}.tar.gz.asc" 2> "${gpg_stderr_device}" | grep -c -E '^\[GNUPG:\] (GOODSIG|VALIDSIG)')
  if [[ ${num_goodlines_tar} -lt 2 ]]; then
    echo "Unable to verify the signature of xt-${TAG}-${OS}-${ARCH}.tar.gz!"
    echo -e "${error_text}"
    exit 1
  fi
  echo "Done."
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    if [[ -n "$INPUT_ARGUMENTS" ]]; then
      echo "Failed to install $BINARY_NAME with the arguments provided: $INPUT_ARGUMENTS"
      help
    else
      echo "Failed to install $BINARY_NAME"
    fi
    echo -e "\tFor support, go to https://github.com/adamkobi/xt."
  fi
  cleanup
  exit $result
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  XT="$(command -v $BINARY_NAME)"
  if [ "$?" = "1" ]; then
    echo "$BINARY_NAME not found. Is $XT_INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
}

# help provides possible cli installation arguments
help () {
  echo "Accepted cli arguments are:"
  echo -e "\t[--help|-h ] ->> prints this help"
  echo -e "\t[--version|-v <desired_version>] . When not defined it fetches the latest release from GitHub"
  echo -e "\te.g. --version v3.0.0 or -v canary"
  echo -e "\t[--no-sudo]  ->> install without sudo"
}

cleanup() {
  if [[ -d "${XT_TMP_ROOT:-}" ]]; then
    rm -rf "$XT_TMP_ROOT"
  fi
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

# Set debug if desired
if [ "${DEBUG}" == "true" ]; then
  set -x
fi

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="${@}"
set -u
while [[ $# -gt 0 ]]; do
  case $1 in
    '--version'|-v)
       shift
       if [[ $# -ne 0 ]]; then
           export DESIRED_VERSION="${1}"
       else
           echo -e "Please provide the desired version. e.g. --version v3.0.0 or -v canary"
           exit 0
       fi
       ;;
    '--no-sudo')
       USE_SUDO="false"
       ;;
    '--help'|-h)
       help
       exit 0
       ;;
    *) exit 1
       ;;
  esac
  shift
done
set +u

initArch
initOS
verifySupported
checkDesiredVersion
if ! checkXtInstalledVersion; then
  downloadFile
  verifyFile
  installFile
fi
testVersion
cleanup