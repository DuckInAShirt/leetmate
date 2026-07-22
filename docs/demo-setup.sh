#!/usr/bin/env bash
# Source this file from the VHS shell so its environment and cleanup trap stay
# active for the whole recording.
set -euo pipefail
umask 077

if [[ ! -x ./leetmate ]]; then
  printf '%s\n' "demo setup: build ./leetmate first" >&2
  return 1
fi

resolved_config="$(./leetmate config)"
source_config_dir="$(awk '/^dir:/ { sub(/^dir:[[:space:]]*/, ""); print; exit }' <<<"${resolved_config}")"
workspace="$(awk '/^leetgo:/ { sub(/^leetgo:[[:space:]]*/, ""); sub(/[[:space:]]+\([^()]*\)$/, ""); print; exit }' <<<"${resolved_config}")"

if [[ -z "${source_config_dir}" || ! -f "${source_config_dir}/config.yaml" ]]; then
  printf '%s\n' "demo setup: resolved LeetMate config not found" >&2
  return 1
fi
if [[ -z "${workspace}" || ! -f "${workspace}/leetgo.yaml" ]]; then
  printf '%s\n' "demo setup: resolved leetgo workspace not found" >&2
  return 1
fi

demo_root="$(mktemp -d "${TMPDIR:-/tmp}/leetmate-demo.XXXXXX")"
cleanup_demo() {
  rm -rf -- "${demo_root}"
}
trap cleanup_demo EXIT
trap 'exit 1' HUP INT TERM

demo_workspace="${demo_root}/workspace"
demo_config="${demo_root}/config"
mkdir -m 700 "${demo_workspace}" "${demo_config}"

# Copy only declarative configuration. `leetgo pick` creates a fresh problem
# skeleton, so personal solutions and notes never enter the demo workspace.
cp "${workspace}/leetgo.yaml" "${demo_workspace}/leetgo.yaml"
cp "${source_config_dir}/config.yaml" "${demo_config}/config.yaml"
perl -pi -e 's/^author:\s*.*/author: leetmate-demo/' "${demo_workspace}/leetgo.yaml"

# Let LeetMate and leetgo parse dotenv files themselves. Symlinks avoid copying
# secrets and live only inside the private, trap-managed temporary directory.
if [[ -f "${source_config_dir}/.env" ]]; then
  ln -s "${source_config_dir}/.env" "${demo_config}/.env"
fi
if [[ -f "${workspace}/.env" ]]; then
  ln -s "${workspace}/.env" "${demo_workspace}/.env"
fi

LEETMATE_CONFIG_DIR="${demo_config}" ./leetmate config set leetgo.workspace "${demo_workspace}" >/dev/null
LEETMATE_CONFIG_DIR="${demo_config}" ./leetmate config set db.path "" >/dev/null

export LEETMATE_CONFIG_DIR="${demo_config}"
export LEETMATE_DEMO_ROOT="${demo_root}"
