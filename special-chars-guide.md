# HTML 特殊字符处理指南

在通过 API 传输 HTML 内容时，特殊字符可能会导致问题。本指南提供了几种处理方案。

## 方案 1：Base64 编码（推荐）

使用 `html_base64` 字段传输 base64 编码的 HTML，可以完全避免特殊字符问题。

### 示例

#### JavaScript
```javascript
// 原始 HTML（包含各种特殊字符）
const html = `
<div class="container">
    <h1>特殊字符测试</h1>
    <p>包含引号: "双引号" 和 '单引号'</p>
    <p>包含尖括号: <tag> 和 </tag></p>
    <p>包含反斜杠: C:\\Windows\\System32</p>
    <p>包含换行符:
        第一行
        第二行
    </p>
    <script>
        console.log("JavaScript 代码");
    </script>
</div>
`;

// 转换为 base64
const htmlBase64 = btoa(unescape(encodeURIComponent(html)));

// 发送请求
fetch('http://localhost:8080/convert', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        html_base64: htmlBase64,
        use_local_server: true
    })
});
```

#### Python
```python
import base64
import requests

html = """
<div class="container">
    <h1>特殊字符测试</h1>
    <p>包含引号: "双引号" 和 '单引号'</p>
    <p>包含尖括号: <tag> 和 </tag></p>
    <p>包含反斜杠: C:\\Windows\\System32</p>
    <script>
        console.log("JavaScript 代码");
    </script>
</div>
"""

# 转换为 base64
html_base64 = base64.b64encode(html.encode('utf-8')).decode('utf-8')

# 发送请求
response = requests.post('http://localhost:8080/convert', json={
    'html_base64': html_base64,
    'use_local_server': True
})
```

#### Shell (bash)
```bash
# 将 HTML 保存到文件
cat > test.html << 'EOF'
<div class="container">
    <h1>特殊字符测试</h1>
    <p>包含引号: "双引号" 和 '单引号'</p>
    <p>包含 $变量 和 `命令`</p>
</div>
EOF

# 转换为 base64 并发送请求
html_base64=$(base64 -w 0 test.html)  # Linux
# html_base64=$(base64 -b 0 test.html)  # macOS

curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d "{\"html_base64\": \"$html_base64\"}"
```

## 方案 2：JSON 字符串转义

如果使用 `html` 字段，需要正确转义 JSON 特殊字符。

### 需要转义的字符

| 字符 | 转义后 | 说明 |
|------|--------|------|
| `"` | `\"` | 双引号 |
| `\` | `\\` | 反斜杠 |
| `/` | `\/` | 斜杠（可选） |
| `\n` | `\\n` | 换行符 |
| `\r` | `\\r` | 回车符 |
| `\t` | `\\t` | 制表符 |

### 示例

#### JavaScript（自动处理）
```javascript
// JavaScript 的 JSON.stringify 会自动处理转义
const html = `<div class="test">包含"引号"的内容</div>`;

fetch('http://localhost:8080/convert', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        html: html  // 自动转义
    })
});
```

#### Python（自动处理）
```python
import json
import requests

html = '<div class="test">包含"引号"的内容</div>'

# Python 的 json 库会自动处理转义
response = requests.post('http://localhost:8080/convert', 
    json={'html': html}  # 自动转义
)
```

#### 手动构建 JSON（不推荐）
```bash
# 如果必须手动构建 JSON，需要正确转义
curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<div class=\"test\">包含\"引号\"的内容</div>"
  }'
```

## 方案 3：使用 Here Document（Shell）

在 Shell 脚本中，可以使用 Here Document 结合 jq 来处理：

```bash
# 使用 jq 自动处理 JSON 转义
html=$(cat << 'EOF'
<div class="complex">
    <h1>复杂 HTML</h1>
    <p>包含各种特殊字符: "引号" '单引号' <标签> & 符号</p>
    <script>
        var data = {"key": "value"};
        console.log("测试");
    </script>
</div>
EOF
)

# jq 会自动处理所有转义
json_payload=$(jq -n --arg html "$html" '{html: $html, use_local_server: true}')

curl -X POST http://localhost:8080/convert \
  -H "Content-Type: application/json" \
  -d "$json_payload"
```

## 最佳实践建议

1. **推荐使用 Base64 编码**
   - 完全避免转义问题
   - 适合包含大量特殊字符的复杂 HTML
   - 支持二进制内容（如内嵌的图片 data URL）

2. **使用成熟的 JSON 库**
   - 不要手动拼接 JSON 字符串
   - 使用编程语言内置的 JSON 库自动处理转义

3. **测试特殊场景**
   - 包含 `<script>` 标签的 HTML
   - 包含内联 CSS 的 HTML
   - 包含 Unicode 字符（如中文、emoji）的 HTML

4. **错误处理**
   - API 会返回明确的错误信息
   - 检查 `success` 字段判断是否成功

## 完整示例：处理复杂 HTML

```javascript
// 一个包含各种特殊情况的完整示例
const complexHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>复杂示例 & 特殊字符</title>
    <style>
        .quote::before { content: '"'; }
        .quote::after { content: '"'; }
        body { font-family: "Microsoft YaHei", '微软雅黑', sans-serif; }
    </style>
</head>
<body>
    <h1>特殊字符测试 😊</h1>
    <p class="quote">这是一个包含"双引号"和'单引号'的段落</p>
    <p>路径示例: C:\\Users\\Documents\\file.txt</p>
    <p>HTML 实体: &lt;tag&gt; &amp; &quot; &apos;</p>
    <pre>
        预格式化文本
        保留    空格和
            缩进
    </pre>
    <script>
        // JavaScript 代码
        var config = {
            "api": "https://example.com/api",
            "token": "abc\"123\"xyz"
        };
        console.log('配置:', config);
    </script>
</body>
</html>
`;

// 方法 1：使用 base64（推荐）
async function convertWithBase64() {
    const base64HTML = btoa(unescape(encodeURIComponent(complexHTML)));
    
    const response = await fetch('http://localhost:8080/convert', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            html_base64: base64HTML,
            use_local_server: true
        })
    });
    
    const result = await response.json();
    if (result.success) {
        console.log('转换成功，图片大小:', result.data.length);
    }
}

// 方法 2：直接使用 JSON（让库处理转义）
async function convertWithJSON() {
    const response = await fetch('http://localhost:8080/convert', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            html: complexHTML,  // 自动转义
            use_local_server: true
        })
    });
    
    const result = await response.json();
    if (result.success) {
        console.log('转换成功');
    }
}
``` 