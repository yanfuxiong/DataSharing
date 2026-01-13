#!/usr/bin/env bash
# install_hooks.sh
# 将本文件尾部 <<<EOF ... EOF 之间的内容写入 .git/hooks/pre-commit

set -euo pipefail


HOOK_DST=".git/hooks/pre-commit"

# 1. 若目标已存在，先删除旧版本
if [[ -f "$HOOK_DST" ]]; then
  echo "Removing existing $HOOK_DST ..."
  rm -f "$HOOK_DST"
fi

# 2. 确保 .git/hooks 目录存在
mkdir -p "$(dirname "$HOOK_DST")"

# 3. 把本文件尾部 HERE DOCUMENT 内容写入并加可执行权限
cat > "$HOOK_DST" <<'EOF'
#!/usr/bin/env bash
# file: .git/hooks/pre-commit
set -euo pipefail

# Change-Id 与 HEAD 相同 → 直接放行，不做pre-commit检查
[ "$(git log -1 --pretty=%B HEAD | sed -n 's/^Change-Id: //p')" = \
  "$(git interpret-trailers --parse <<<"$(git log -1 --pretty=%B HEAD)" | sed -n 's/^Change-Id: //p')" ] && \
  { echo "DEBUG: Change-Id unchanged, skip pre-commit" >&2; exit 0; }

###############################  用户可配置区  ###########################################
# 配置表：	4个参数段，用 | 分隔
# 参数内容：  目录通配串（空格分隔） | 版本文件 | 版本变量名 	| 白名单子目录（空格分隔，可空）
# 参数含义：  1: 需要监控的目录（空格分隔，支持通配） 
# 			2: 监控目录有变更后需要更新的版本内容所在的版本文件名
# 			3：监控目录有变更后需要修改的版本文件名内的版本变量名称
# 			4：监控目录里要剔除的子目录（空格分隔，支持通配，支持空）
# 功能：
#   1. 检查 参数1 中的目录是否有变更；
#   2. 有变更时，则检查参数2中的 参数3的值是否改变，没改变则告警并中断提交；
#   3. 检查参数1的目录支持白名单功能， 白名单内容为参数4
#   4. 若 FORBID_DIRS 中的子目录有任何变更/新增，直接报错退出。
########################################################################################
CONFIG=$(cat <<'CONFIG_EOF'
client/*   misc/* |client/global/const.go 								|ClientVersion	  |client/platform/macOs/CrossShare_mac/* client/platform/iOS/CrossShare_iOS/*
lanServer/* misc/*|lanServer/global/const.go							|LanServerVersion |
clipboard_java/*  |clipboard_java/app/build.gradle  					|versionName      |
windows/*         |windows/source_code/windows_clipboard/CMakeLists.txt |VERSION_PATCH    |
client/platform/iOS/CrossShare_iOS/*  |client/platform/iOS/CrossShare_iOS/CrossShare_iOS.xcodeproj/project.pbxproj|MARKETING_VERSION|
client/platform/macOs/CrossShare_mac/*|client/platform/macOs/CrossShare_mac/CrossShare.xcodeproj/project.pbxproj  |MARKETING_VERSION|
CONFIG_EOF
)
FORBID_DIRS="client/platform/macOs/build/* client/platform/iOS/build/* client/platform/android/build/* client/platform/windows/build/*"


# 颜色辅助
RED='\033[0;31m'; YELLOW='\033[0;33m'; NC='\033[0m'

#######################  绝对禁止变更的子目录（空格分隔，支持通配） #####################
# 1. 先检查“禁止目录” FORBID_DIRS 是否有变更，有变更则不允许提交
###################################################################################
for fd in $FORBID_DIRS; do
  # 用 git diff --name-only 比较暂存区与 HEAD
  if git diff --cached --name-only --diff-filter=AM HEAD | grep -qE "^$fd(/|$)"; then
    echo -e "${RED}ERROR: The directory: $fd that prohibits changes has been added/modified, and the submission has been rejected!!!${NC}" >&2
    exit 1
  fi
done


#########################################
# 提取版本变量值函数（兼容多种写法）
# 支持：
#   var = "1.2.3"
#   var "1.2.3"
#   var = 1.2.3;
#   set(var 1.2.3)
#   MARKETING_VERSION = 1.0.2;
# 返回 NOT_FOUND 表示未找到
#########################################
function extract_version() {
  local ref="$1" file="$2" var="$3"
  local result

  # 将变量名直接嵌入到 perl 脚本中，避免环境变量传递问题
  result=$(git show "$ref:$file" 2>/dev/null | perl -sne '
    BEGIN { $v = shift }
    # 1) var = "1.2.3"  或 var "1.2.3"
    if (m/^\s*\Q$v\E\s*[=:]?\s*"([^"]+)"/) { print $1; exit }
    # 2) var = 1.2.3;  或 var 1.2.3
    if (m/^\s*\Q$v\E\s*[=:]?\s*([0-9][0-9A-Za-z._-]*)\s*;?/) { print $1; exit }
    # 3) set(VAR 1.2.3)
    if (m/^\s*set\(\s*\Q$v\E\s+([^) \t]+)\s*\)/i) { print $1; exit }
  ' -- "$var")

  if [ -z "$result" ]; then
    echo "NOT_FOUND"
  else
    echo "$result"
  fi
}

#########################################
# 检查配置项功能函数
# $1 监控目录通配串（空格分隔）
# $2 版本文件（相对仓库根）
# $3 版本变量名 
# $4 白名单目录 （空格分隔，支持通配）
#########################################
function check_mod_and_ver() {
  local dir_list=$1 
  local const_file=$(echo "$2"|xargs)
  local var_name=$(echo "$3"|xargs)
  local exclude_dir=$4

  # 清理首尾空格，保持内部空格用于分隔多个路径
  dir_list=$(echo "$dir_list" | xargs)
  exclude_dir=$(echo "$exclude_dir" | xargs)

  # 1. 指定目录变更检查（剔除白名单子目录）
  local paths=()
  if [ -n "$dir_list" ]; then
    local old_ifs=$IFS
    IFS=' '
    # shellcheck disable=SC2206
    paths=($dir_list)
    IFS=$old_ifs
  fi

  local changed=""
  if [ ${#paths[@]} -gt 0 ]; then
    changed=$(git diff --cached --name-only --diff-filter=AM HEAD -- "${paths[@]}")
  fi

  # 白名单过滤（用 perl 前缀匹配，通配符->正则）
  if [ -n "$changed" ] && [ -n "$exclude_dir" ]; then
    # 把白名单通配符转成正则，一行一个
    local regex
    regex=$(echo "$exclude_dir" | tr ' ' '\n' | sed 's/\*/.*/g' | paste -sd'|')
    # perl：只要路径以任一白名单前缀开头就跳过
    changed=$(echo "$changed" | perl -nle "print unless m{^(?:$regex)(/|\$)}")
  fi

  if [ -z "$changed" ]; then 
    # 无变更，直接放行    
    # echo "DEBUG: dirs=[$dir_list] exclude=[$exclude_dir], changed is null!" >&2
    return 0
  fi

  # 2. 有变更，则检查Version 是否同步修改
  # 提取旧值/新值  （兼容 macOS 无 -P，用 perl）
  local head_v stag_v
  head_v=$(extract_version "HEAD" "$const_file" "$var_name")
  stag_v=$(extract_version "" "$const_file" "$var_name")
  #echo "DEBUG: head_version=[$head_v] stag_version=[$stag_v]" >&2
  if [ "$head_v" = "$stag_v" ]; then
    echo -e "${YELLOW}WARNING: Detected changes in [$dir_list] but $const_file::$var_name not updated!${NC}" >&2
    echo "$changed" >&2
    echo -e "${YELLOW}Please check and modify the value of $const_file::$var_name before submitting!!! ${NC}" >&2
    exit 1     # 变量没有变更 → 阻断提交
  else
    echo "check: dirs=[$dir_list] exclude=[$exclude_dir], check success!" >&2
    return 0
  fi
}


#################################################
# 2. 主循环：逐行读取配置表并执行检查
#################################################
while IFS='|' read -r dirs const_file vars exclude; do
  # 跳过空行/注释
  [[ -z "$dirs" || "$dirs" =~ ^[[:space:]]*# ]] && continue
  check_mod_and_ver "$dirs" "$const_file" "$vars" "$exclude"
done <<< "$CONFIG"

EOF

# 4. 加可执行权限
chmod +x "$HOOK_DST"

echo "Git pre-commit hook installed successfully → $HOOK_DST"