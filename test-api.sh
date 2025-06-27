#!/bin/bash

# HTML2Image API 测试脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# API 地址
API_URL="${API_URL:-http://localhost:8080}"

echo -e "${GREEN}=== HTML2Image API 测试 ===${NC}"
echo -e "API 地址: ${API_URL}\n"

# 测试健康检查
echo -e "${YELLOW}1. 测试健康检查端点...${NC}"
health_response=$(curl -s "${API_URL}/health")
if echo "$health_response" | grep -q '"healthy":true'; then
    echo -e "${GREEN}✓ 健康检查通过${NC}"
else
    echo -e "${RED}✗ 健康检查失败${NC}"
    echo "响应: $health_response"
fi
echo

# 测试服务信息
echo -e "${YELLOW}2. 获取服务信息...${NC}"
info_response=$(curl -s "${API_URL}/")
echo "$info_response" | jq .
echo

# 测试简单 HTML 转换
echo -e "${YELLOW}3. 测试简单 HTML 转换...${NC}"
simple_html='<html><body style="padding: 20px;"><h1>测试标题</h1><p>这是一个测试段落。</p></body></html>'
convert_response=$(curl -s -X POST "${API_URL}/convert" \
    -H "Content-Type: application/json" \
    -d "{\"html\": \"$simple_html\"}")

if echo "$convert_response" | jq -e '.success' > /dev/null; then
    echo -e "${GREEN}✓ 转换成功${NC}"
    # 保存图片
    echo "$convert_response" | jq -r '.data' | base64 -d > test-simple.png
    echo "图片已保存到: test-simple.png"
else
    echo -e "${RED}✗ 转换失败${NC}"
    echo "$convert_response" | jq .
fi
echo

# 测试带外部图片的 HTML
echo -e "${YELLOW}4. 测试带外部图片的 HTML（使用本地服务器模式）...${NC}"
complex_html=$(cat << 'EOF'
<html>
<head>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            padding: 20px; 
            background-color: #f0f0f0;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { color: #333; }
        img { margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>带外部图片的测试</h1>
        <p>下面是一个来自外部的图片：</p>
        <img src="https://via.placeholder.com/300x200" alt="测试图片">
        <p>如果能看到上面的图片，说明外部资源加载成功。</p>
    </div>
</body>
</html>
EOF
)

# 转义 HTML 中的特殊字符
escaped_html=$(echo "$complex_html" | jq -Rs .)

convert_response=$(curl -s -X POST "${API_URL}/convert" \
    -H "Content-Type: application/json" \
    -d "{\"html\": $escaped_html, \"use_local_server\": true}")

if echo "$convert_response" | jq -e '.success' > /dev/null; then
    echo -e "${GREEN}✓ 转换成功${NC}"
    # 保存图片
    echo "$convert_response" | jq -r '.data' | base64 -d > test-complex.png
    echo "图片已保存到: test-complex.png"
else
    echo -e "${RED}✗ 转换失败${NC}"
    echo "$convert_response" | jq .
fi
echo

# 测试中文内容
echo -e "${YELLOW}5. 测试中文内容...${NC}"
chinese_html=$(cat << 'EOF'
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { 
            font-family: "Microsoft YaHei", "微软雅黑", Arial, sans-serif;
            padding: 20px;
        }
    </style>
</head>
<body>
    <h1>中文标题测试</h1>
    <p>这是一段中文内容，用于测试中文字体的显示效果。</p>
    <ul>
        <li>列表项目一</li>
        <li>列表项目二</li>
        <li>列表项目三</li>
    </ul>
</body>
</html>
EOF
)

escaped_html=$(echo "$chinese_html" | jq -Rs .)

convert_response=$(curl -s -X POST "${API_URL}/convert" \
    -H "Content-Type: application/json" \
    -d "{\"html\": $escaped_html}")

if echo "$convert_response" | jq -e '.success' > /dev/null; then
    echo -e "${GREEN}✓ 转换成功${NC}"
    # 保存图片
    echo "$convert_response" | jq -r '.data' | base64 -d > test-chinese.png
    echo "图片已保存到: test-chinese.png"
else
    echo -e "${RED}✗ 转换失败${NC}"
    echo "$convert_response" | jq .
fi

echo

# 测试 Base64 编码的 HTML
echo -e "${YELLOW}6. 测试 Base64 编码的 HTML（包含特殊字符）...${NC}"
special_html='<div class="special">
    <h1>特殊字符测试</h1>
    <p>包含"双引号"和'\''单引号'\''</p>
    <p>包含<标签>和&符号</p>
    <script>console.log("JS代码");</script>
</div>'

# 转换为 base64
html_base64=$(echo -n "$special_html" | base64)

convert_response=$(curl -s -X POST "${API_URL}/convert" \
    -H "Content-Type: application/json" \
    -d "{\"html_base64\": \"$html_base64\"}")

if echo "$convert_response" | jq -e '.success' > /dev/null; then
    echo -e "${GREEN}✓ Base64 转换成功${NC}"
    # 保存图片
    echo "$convert_response" | jq -r '.data' | base64 -d > test-base64.png
    echo "图片已保存到: test-base64.png"
else
    echo -e "${RED}✗ Base64 转换失败${NC}"
    echo "$convert_response" | jq .
fi

echo -e "\n${GREEN}=== 测试完成 ===${NC}"
echo -e "生成的图片文件："
ls -la test-*.png 2>/dev/null || echo "没有生成图片文件" 