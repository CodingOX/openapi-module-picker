#!/usr/bin/env bash
set -euo pipefail

BINARY="openapi-trim"
TARGET="/usr/local/bin"

echo "==> 构建 ${BINARY} ..."
go build -o "${BINARY}" .

echo "==> 安装到 ${TARGET}/${BINARY} ..."
sudo cp "${BINARY}" "${TARGET}/"

echo "==> 清理临时二进制 ..."
rm -f "${BINARY}"

echo ""
echo "✅ 安装完成！运行以下命令验证："
echo "   ${BINARY}"
