#!/bin/sh
echo "=== Preparing Kubeconfig for Docker Container ==="

if [ -f "$KUBECONFIG" ]; then
    echo "✅ Found kubeconfig at $KUBECONFIG"
    echo "📄 Original content (server lines):"
    grep -n "server:" "$KUBECONFIG" || echo "No server: found"
    
    # 使用 awk 或 sed 进行替换（更可靠的方法）
    # 处理可能的 base64 编码内容（certificates 等）
    
    # 方法 1: 直接 sed 替换
    if command -v sed >/dev/null 2>&1; then
        echo "🔧 Attempting sed replacement..."
        sed -i 's|https://127\.0\.0\.1:|https://host.docker.internal:|g' "$KUBECONFIG"
        sed -i 's|http://127\.0\.0\.1:|http://host.docker.internal:|g' "$KUBECONFIG"
        sed -i 's|127\.0\.0\.1:[0-9]*|host.docker.internal|g' "$KUBECONFIG"
    fi
    
    echo ""
    echo "📄 Updated content (server lines):"
    grep -n "server:" "$KUBECONFIG" || echo "No server: found"
    echo ""
    echo "✅ Kubeconfig processing completed"
else
    echo "❌ Kubeconfig not found at $KUBECONFIG"
    echo "📁 Listing /home/nonroot/.kube/:"
    ls -la /home/nonroot/.kube/ 2>/dev/null || echo "Directory not found"
fi

# Start the actual service
exec /usr/local/bin/cluster-service
