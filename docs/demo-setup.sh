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
# 录 demo 专用：换上 SiliconFlow 上最快的 instruct 模型（Qwen3-30B-A3B，MoE 仅激活 3B，
# 实测 ~0.5s 响应，比默认 DeepSeek-V4-Flash 快约 18 倍），让 Hint/Nudge/Review 在镜头内流完。
# 仅覆盖 model，preset/provider 仍是 siliconflow，复用 SILICONFLOW_API_KEY。
LEETMATE_CONFIG_DIR="${demo_config}" ./leetmate config set llm.model Qwen/Qwen3-30B-A3B-Instruct-2507 >/dev/null

# Pre-pick the first Hot 100 problem (two-sum) and seed an EMPTY function body
# (signature only), so the demo can show a real "write code from scratch" moment
# by typing the whole hash-map solution in the built-in editor.
#
# We also strip the `# @lc code=end` marker and the local `if __name__` test
# block: the built-in editor opens with the cursor at the file's end, so to make
# the typed lines land INSIDE leetgo's submit region (`@lc code=begin` .. end),
# the file must end right after the signature. leetgo still submits `begin`→EOF
# and accepts the solution. leetgo pick never overwrites an existing solution,
# so the seeded skeleton survives the demo's own pick.
( cd "${demo_workspace}" && leetgo pick 1 >/dev/null 2>&1 ) || true
seed_sol="${demo_workspace}/python/0001.two-sum/solution.py"
if [ -f "${seed_sol}" ]; then
  python3 - "${seed_sol}" <<'PYEOF'
import sys, re
p = sys.argv[1]; s = open(p).read()
pat = r'(    def twoSum\(self, nums: List\[int\], target: int\) -> List\[int\]:)\n[ \t]*\n'
rep = r'\1\n'
s2, n = re.subn(pat, rep, s)
s2 = re.sub(r'\n# @lc code=end.*', '', s2, flags=re.DOTALL)
open(p, 'w').write(s2 if n else s)
PYEOF
fi

export LEETMATE_CONFIG_DIR="${demo_config}"
export LEETMATE_DEMO_ROOT="${demo_root}"
